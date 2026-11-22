package consts

const (
	DefaultServerConfigPath        = "configs/kubeconfig.yaml"
	DefaultServerNamespace         = "kube-buding"
	GroupCrd                string = "kube.buding.goaigc.fun"
	ClusterCrd              string = "clusters.kube.buding.goaigc.fun"
	ClusterKind                    = "Cluster"

	UserCrd           string = "users.kube.buding.goaigc.fun"
	UserKind                 = "User"
	ClusterCrdVersion string = "v1"
)

// StorageClass 相关注解
const (
	// StorageClassTypeAnnotation 手动设置的存储类型: block / file / object; 未设置时按供应者自动识别
	StorageClassTypeAnnotation = GroupCrd + "/storage-type"
	// StorageClassDefaultAnnotation 默认存储类(Kubernetes 标准注解)
	StorageClassDefaultAnnotation     = "storageclass.kubernetes.io/is-default-class"
	StorageClassDefaultBetaAnnotation = "storageclass.beta.kubernetes.io/is-default-class"
	// 以下两个注解与 KubeSphere 保持一致, 两边的界面可以互相识别
	StorageClassAllowCloneAnnotation  = "storageclass.kubesphere.io/allow-clone"
	StorageClassAccessModesAnnotation = "storageclass.kubesphere.io/supported-access-modes"
	// KubeSphere 的别名 / 描述注解, 读取时作为兜底
	KubeSphereAliasAnnotation       = "kubesphere.io/alias-name"
	KubeSphereDescriptionAnnotation = "kubesphere.io/description"
)

// 通用注解
const (
	// CreatorAnnotation 创建者(本平台创建资源时写入当前登录用户)
	CreatorAnnotation = GroupCrd + "/creator"
	// KubeSphereCreatorAnnotation KubeSphere 的创建者注解, 读取时作为兜底
	KubeSphereCreatorAnnotation = "kubesphere.io/creator"
	// JobRunRecordsAnnotation 任务重新执行前的运行记录(JSON 数组)
	JobRunRecordsAnnotation = GroupCrd + "/run-records"
	// CronJobInstantiateAnnotation 值为 manual 表示任务是手动触发的(与 kubectl create job --from=cronjob 相同)
	CronJobInstantiateAnnotation = "cronjob.kubernetes.io/instantiate"
)
