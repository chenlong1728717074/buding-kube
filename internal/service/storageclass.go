package service

import (
	"buding-kube/internal/web/dto"
	"buding-kube/internal/web/vo"
	"buding-kube/pkg/consts"
	"buding-kube/pkg/logs"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/yaml"
)

var (
	storageClassSrv  *StorageClassService
	storageClassOnce sync.Once
)

type StorageClassService struct {
}

func NewStorageClassService() *StorageClassService {
	return &StorageClassService{}
}

func GetSingletonStorageClassService() *StorageClassService {
	storageClassOnce.Do(func() {
		storageClassSrv = NewStorageClassService()
	})
	return storageClassSrv
}

// 存储类型标识与中文名
const (
	storageTypeBlock  = "block"
	storageTypeFile   = "file"
	storageTypeObject = "object"
	storageTypeNone   = "未分类"
)

var storageTypeLabels = map[string]string{
	storageTypeBlock:  "块存储",
	storageTypeFile:   "文件存储",
	storageTypeObject: "对象存储",
}

func storageTypeLabel(key string) string {
	if label, ok := storageTypeLabels[key]; ok {
		return label
	}
	return storageTypeNone
}

// normalizeStorageType 把 block / 块存储 等写法统一成标识; 空字符串表示自动识别
func normalizeStorageType(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" || value == "auto" {
		return "", nil
	}
	if _, ok := storageTypeLabels[value]; ok {
		return value, nil
	}
	for key, label := range storageTypeLabels {
		if label == value {
			return key, nil
		}
	}
	return "", fmt.Errorf("不支持的存储类型: %s, 可选 block / file / object", value)
}

// classifyStorageType 按供应者推断存储类型, 无法识别时返回空
func classifyStorageType(provisioner string) string {
	p := strings.ToLower(provisioner)
	switch {
	// 文件存储
	case strings.Contains(p, "nfs"), strings.Contains(p, "cephfs"), strings.Contains(p, "glusterfs"),
		strings.Contains(p, "nas"), strings.Contains(p, "cpfs"), strings.Contains(p, "filestore"),
		strings.Contains(p, "azurefile"), strings.Contains(p, "efs"), strings.Contains(p, "juicefs"),
		strings.Contains(p, "jfs"):
		return storageTypeFile
	// 对象存储
	case strings.Contains(p, "oss"), strings.Contains(p, "s3"), strings.Contains(p, "obs"),
		strings.Contains(p, "cos"), strings.Contains(p, "minio"), strings.Contains(p, "object"),
		strings.Contains(p, "bos"):
		return storageTypeObject
	// 块存储
	case strings.Contains(p, "disk"), strings.Contains(p, "rbd"), strings.Contains(p, "block"),
		strings.Contains(p, "ebs"), strings.Contains(p, "longhorn"), strings.Contains(p, "openebs"),
		strings.Contains(p, "rook"), strings.Contains(p, "cbs"), strings.Contains(p, "csi"),
		strings.Contains(p, "local"), strings.Contains(p, "no-provisioner"):
		return storageTypeBlock
	default:
		return ""
	}
}

// resolveStorageType 优先使用手动设置的存储类型, 否则按供应者推断; 第二个返回值表示是否为自动识别
func resolveStorageType(sc *storagev1.StorageClass) (string, bool) {
	if key, err := normalizeStorageType(sc.Annotations[consts.StorageClassTypeAnnotation]); err == nil && key != "" {
		return key, false
	}
	return classifyStorageType(sc.Provisioner), true
}

// provisionerSupportsClone 基于供应者推断是否支持卷克隆(常见 CSI/存储驱动支持 DataSource 克隆)
func provisionerSupportsClone(provisioner string) bool {
	p := strings.ToLower(provisioner)
	switch {
	case strings.Contains(p, "csi"), strings.Contains(p, "rbd"), strings.Contains(p, "rook"),
		strings.Contains(p, "ceph"), strings.Contains(p, "ebs"), strings.Contains(p, "longhorn"),
		strings.Contains(p, "openebs"):
		return true
	default:
		return false
	}
}

// resolveAllowClone 优先使用注解中的设置, 否则按供应者推断
func resolveAllowClone(sc *storagev1.StorageClass) bool {
	switch sc.Annotations[consts.StorageClassAllowCloneAnnotation] {
	case "true":
		return true
	case "false":
		return false
	}
	return provisionerSupportsClone(sc.Provisioner)
}

func isDefaultStorageClass(sc *storagev1.StorageClass) bool {
	return sc.Annotations[consts.StorageClassDefaultAnnotation] == "true" ||
		sc.Annotations[consts.StorageClassDefaultBetaAnnotation] == "true"
}

var validAccessModes = map[string]bool{
	string(corev1.ReadWriteOnce):    true,
	string(corev1.ReadOnlyMany):     true,
	string(corev1.ReadWriteMany):    true,
	string(corev1.ReadWriteOncePod): true,
}

func (s *StorageClassService) List(query dto.StorageClassPageQueryBaseDTO) ([]vo.StorageClassVO, error) {
	clientSet, err := ClusterMap.Get(query.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", query.ClusterId, err.Error())
		return nil, errors.New("获取集群失败")
	}
	items, err := clientSet.StorageV1().StorageClasses().List(context.TODO(), metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		logs.Error("获取StorageClass失败: %v", err)
		return nil, fmt.Errorf("获取StorageClass失败: %v", err)
	}

	// 统计每个存储类关联的持久卷声明(PVC)数量
	pvcCountMap := make(map[string]int64)
	if pvcs, err := clientSet.CoreV1().PersistentVolumeClaims(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{ResourceVersion: "0"}); err == nil {
		for _, pvc := range pvcs.Items {
			if pvc.Spec.StorageClassName == nil {
				// 未显式指定时使用集群默认存储类, 不计入具体存储类
				continue
			}
			pvcCountMap[*pvc.Spec.StorageClassName]++
		}
	} else {
		logs.Error("获取PVC列表失败, 持久卷声明数量统计跳过: %v", err)
	}

	typeFilter := strings.TrimSpace(query.StorageType)
	result := make([]vo.StorageClassVO, 0)
	for i := range items.Items {
		item := &items.Items[i]
		typeKey, _ := resolveStorageType(item)
		typeLabel := storageTypeLabel(typeKey)
		if typeFilter != "" && typeFilter != typeKey && typeFilter != typeLabel {
			continue
		}
		if query.Keyword != "" && !strings.Contains(item.Name, query.Keyword) {
			continue
		}
		result = append(result, storageClassToVO(item, pvcCountMap[item.Name]))
	}
	// 默认存储类排在最前, 其余按创建时间倒序
	sort.SliceStable(result, func(i, j int) bool {
		if result[i].IsDefault != result[j].IsDefault {
			return result[i].IsDefault
		}
		return result[i].CreateTime.After(result[j].CreateTime)
	})
	return result, nil
}

// storageClassToVO 生成列表 / 详情共用的展示数据
func storageClassToVO(item *storagev1.StorageClass, pvcCount int64) vo.StorageClassVO {
	typeKey, auto := resolveStorageType(item)
	vi := vo.StorageClass2VO(*item)
	vi.StorageType = storageTypeLabel(typeKey)
	vi.StorageTypeKey = typeKey
	vi.StorageTypeAuto = auto
	vi.PVCCount = pvcCount
	vi.AllowVolumeClone = resolveAllowClone(item)
	return vi
}

// Get 存储类详情, 包含使用该存储类的持久卷声明
func (s *StorageClassService) Get(ctx context.Context, base dto.StorageClassBaseDTO) (*vo.StorageClassDetailVO, error) {
	name := strings.TrimSpace(base.Name)
	if name == "" {
		return nil, errors.New("存储类名称不能为空")
	}
	clientSet, err := ClusterMap.GetClientSet(base.ClusterId)
	if err != nil {
		return nil, err
	}
	sc, err := clientSet.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("存储类 %s 不存在", name)
		}
		return nil, fmt.Errorf("获取存储类失败: %v", err)
	}

	pvcs, err := clientSet.CoreV1().PersistentVolumeClaims(metav1.NamespaceAll).List(ctx, metav1.ListOptions{ResourceVersion: "0"})
	if err != nil {
		return nil, fmt.Errorf("获取持久卷声明失败: %v", err)
	}
	claims := make([]vo.PVCVO, 0)
	for i := range pvcs.Items {
		if c := pvcs.Items[i].Spec.StorageClassName; c != nil && *c == name {
			claims = append(claims, vo.PVC2VO(pvcs.Items[i]))
		}
	}
	sort.Slice(claims, func(i, j int) bool {
		return claims[i].CreateTime.After(claims[j].CreateTime)
	})

	var pvCount int64
	if pvs, err := clientSet.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{ResourceVersion: "0"}); err == nil {
		for i := range pvs.Items {
			if pvs.Items[i].Spec.StorageClassName == name {
				pvCount++
			}
		}
	} else {
		logs.Warn("获取持久卷失败, 持久卷数量统计跳过: %v", err)
	}

	return &vo.StorageClassDetailVO{
		StorageClassVO:         storageClassToVO(sc, int64(len(claims))),
		PersistentVolumeClaims: claims,
		PVCount:                pvCount,
	}, nil
}

// Create 创建存储类
func (s *StorageClassService) Create(ctx context.Context, req dto.StorageClassCreateDTO) error {
	clientSet, err := ClusterMap.GetClientSet(req.ClusterId)
	if err != nil {
		return err
	}
	sc, err := buildStorageClass(req)
	if err != nil {
		return err
	}
	if _, err = clientSet.StorageV1().StorageClasses().Create(ctx, sc, metav1.CreateOptions{}); err != nil {
		if apierrors.IsAlreadyExists(err) {
			return fmt.Errorf("存储类 %s 已存在", sc.Name)
		}
		logs.Error("创建StorageClass失败: %v", err)
		return fmt.Errorf("创建存储类失败: %v", err)
	}
	if req.IsDefault {
		s.unsetOtherDefaults(ctx, clientSet, sc.Name)
	}
	logs.Info("StorageClass %s 创建成功", sc.Name)
	return nil
}

func buildStorageClass(req dto.StorageClassCreateDTO) (*storagev1.StorageClass, error) {
	name := strings.TrimSpace(req.Name)
	if errs := validation.IsDNS1123Subdomain(name); len(errs) > 0 {
		return nil, fmt.Errorf("名称不合法: %s", strings.Join(errs, "; "))
	}
	provisioner := strings.TrimSpace(req.Provisioner)
	if provisioner == "" {
		return nil, errors.New("供应者不能为空")
	}
	storageType, err := normalizeStorageType(req.StorageType)
	if err != nil {
		return nil, err
	}

	reclaim := corev1.PersistentVolumeReclaimPolicy(strings.TrimSpace(req.ReclaimPolicy))
	switch reclaim {
	case "":
		reclaim = corev1.PersistentVolumeReclaimDelete
	case corev1.PersistentVolumeReclaimDelete, corev1.PersistentVolumeReclaimRetain:
	default:
		return nil, fmt.Errorf("不支持的回收策略: %s, 可选 Delete / Retain", reclaim)
	}

	bindingMode := storagev1.VolumeBindingMode(strings.TrimSpace(req.VolumeBindingMode))
	switch bindingMode {
	case "":
		bindingMode = storagev1.VolumeBindingImmediate
	case storagev1.VolumeBindingImmediate, storagev1.VolumeBindingWaitForFirstConsumer:
	default:
		return nil, fmt.Errorf("不支持的卷绑定模式: %s, 可选 Immediate / WaitForFirstConsumer", bindingMode)
	}

	params := make(map[string]string, len(req.Parameters))
	for k, v := range req.Parameters {
		if k = strings.TrimSpace(k); k != "" {
			params[k] = v
		}
	}

	annotations := map[string]string{}
	setOrDelete(annotations, "alias", strings.TrimSpace(req.Alias))
	setOrDelete(annotations, "describe", strings.TrimSpace(req.Describe))
	setOrDelete(annotations, consts.StorageClassTypeAnnotation, storageType)
	annotations[consts.StorageClassAllowCloneAnnotation] = fmt.Sprintf("%t", req.AllowVolumeClone)
	if len(req.AccessModes) > 0 {
		for _, mode := range req.AccessModes {
			if !validAccessModes[mode] {
				return nil, fmt.Errorf("不支持的访问模式: %s", mode)
			}
		}
		raw, _ := json.Marshal(req.AccessModes)
		annotations[consts.StorageClassAccessModesAnnotation] = string(raw)
	}
	if req.IsDefault {
		annotations[consts.StorageClassDefaultAnnotation] = "true"
	}

	mountOptions := make([]string, 0, len(req.MountOptions))
	for _, opt := range req.MountOptions {
		if opt = strings.TrimSpace(opt); opt != "" {
			mountOptions = append(mountOptions, opt)
		}
	}

	allowExpansion := req.AllowVolumeExpansion
	sc := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Annotations: annotations,
		},
		Provisioner:          provisioner,
		Parameters:           params,
		ReclaimPolicy:        &reclaim,
		VolumeBindingMode:    &bindingMode,
		AllowVolumeExpansion: &allowExpansion,
	}
	if len(mountOptions) > 0 {
		sc.MountOptions = mountOptions
	}
	return sc, nil
}

// Update 编辑存储类的别名、描述、存储类型、卷扩展、卷克隆和默认存储类
func (s *StorageClassService) Update(ctx context.Context, req dto.StorageClassUpdateDTO) error {
	clientSet, err := ClusterMap.GetClientSet(req.ClusterId)
	if err != nil {
		return err
	}
	sc, err := clientSet.StorageV1().StorageClasses().Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return fmt.Errorf("存储类 %s 不存在", req.Name)
		}
		return fmt.Errorf("获取存储类失败: %v", err)
	}
	storageType, err := normalizeStorageType(req.StorageType)
	if err != nil {
		return err
	}

	if sc.Annotations == nil {
		sc.Annotations = map[string]string{}
	}
	setOrDelete(sc.Annotations, "alias", strings.TrimSpace(req.Alias))
	setOrDelete(sc.Annotations, "describe", strings.TrimSpace(req.Describe))
	// 统一写入本系统的注解后, 去掉 KubeSphere 的同义注解, 避免两边显示不一致
	delete(sc.Annotations, consts.KubeSphereAliasAnnotation)
	delete(sc.Annotations, consts.KubeSphereDescriptionAnnotation)
	setOrDelete(sc.Annotations, consts.StorageClassTypeAnnotation, storageType)
	if req.AllowVolumeClone != nil {
		sc.Annotations[consts.StorageClassAllowCloneAnnotation] = fmt.Sprintf("%t", *req.AllowVolumeClone)
	}
	if req.AllowVolumeExpansion != nil {
		allow := *req.AllowVolumeExpansion
		sc.AllowVolumeExpansion = &allow
	}
	if req.IsDefault != nil {
		if *req.IsDefault {
			sc.Annotations[consts.StorageClassDefaultAnnotation] = "true"
		} else {
			delete(sc.Annotations, consts.StorageClassDefaultAnnotation)
		}
		delete(sc.Annotations, consts.StorageClassDefaultBetaAnnotation)
	}

	if _, err = clientSet.StorageV1().StorageClasses().Update(ctx, sc, metav1.UpdateOptions{}); err != nil {
		logs.Error("更新StorageClass失败: %v", err)
		return fmt.Errorf("更新存储类失败: %v", err)
	}
	if req.IsDefault != nil && *req.IsDefault {
		s.unsetOtherDefaults(ctx, clientSet, sc.Name)
	}
	logs.Info("StorageClass %s 更新成功", sc.Name)
	return nil
}

// Apply 通过 YAML 创建或更新存储类
func (s *StorageClassService) Apply(ctx context.Context, req dto.StorageClassApplyDTO) error {
	clientSet, err := ClusterMap.GetClientSet(req.ClusterId)
	if err != nil {
		return err
	}
	var sc storagev1.StorageClass
	if err = yaml.Unmarshal([]byte(req.Yaml), &sc); err != nil {
		return fmt.Errorf("解析YAML失败: %v", err)
	}
	if sc.Kind != "" && sc.Kind != "StorageClass" {
		return fmt.Errorf("YAML 的 kind 应为 StorageClass, 实际为 %s", sc.Kind)
	}
	if sc.Name == "" {
		return errors.New("metadata.name 不能为空")
	}
	if sc.Provisioner == "" {
		return errors.New("provisioner 不能为空")
	}

	existing, err := clientSet.StorageV1().StorageClasses().Get(ctx, sc.Name, metav1.GetOptions{})
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("获取存储类失败: %v", err)
		}
		sc.ResourceVersion = ""
		if _, err = clientSet.StorageV1().StorageClasses().Create(ctx, &sc, metav1.CreateOptions{}); err != nil {
			return fmt.Errorf("创建存储类失败: %v", err)
		}
	} else {
		// 供应者、参数、回收策略、绑定模式不可修改, 修改这些字段时 apiserver 会直接返回错误
		sc.ResourceVersion = existing.ResourceVersion
		if _, err = clientSet.StorageV1().StorageClasses().Update(ctx, &sc, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("更新存储类失败: %v", err)
		}
	}
	if isDefaultStorageClass(&sc) {
		s.unsetOtherDefaults(ctx, clientSet, sc.Name)
	}
	return nil
}

// unsetOtherDefaults 设置默认存储类后, 取消其他存储类的默认标记, 保证集群只有一个默认存储类
func (s *StorageClassService) unsetOtherDefaults(ctx context.Context, clientSet *kubernetes.Clientset, keep string) {
	list, err := clientSet.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		logs.Warn("取消其他默认存储类失败: %v", err)
		return
	}
	for i := range list.Items {
		other := &list.Items[i]
		if other.Name == keep || !isDefaultStorageClass(other) {
			continue
		}
		delete(other.Annotations, consts.StorageClassDefaultAnnotation)
		delete(other.Annotations, consts.StorageClassDefaultBetaAnnotation)
		if _, err = clientSet.StorageV1().StorageClasses().Update(ctx, other, metav1.UpdateOptions{}); err != nil {
			logs.Warn("取消存储类 %s 的默认标记失败: %v", other.Name, err)
		}
	}
}

func setOrDelete(m map[string]string, key, value string) {
	if value == "" {
		delete(m, key)
		return
	}
	m[key] = value
}

// Delete 删除StorageClass
func (s *StorageClassService) Delete(base dto.BaseDTO) error {
	clientSet, err := ClusterMap.Get(base.ClusterId)
	if err != nil {
		logs.Error("获取集群失败: %s %s", base.ClusterId, err.Error())
		return errors.New("获取集群失败")
	}
	if err := clientSet.StorageV1().StorageClasses().Delete(context.TODO(), base.Name, metav1.DeleteOptions{}); err != nil {
		logs.Error("删除storageclass失败: %v", err)
		return fmt.Errorf("删除storageclass失败: %v", err)
	}
	return nil
}
