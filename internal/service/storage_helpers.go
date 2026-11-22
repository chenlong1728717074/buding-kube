package service

import (
	"errors"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// parseStorageQuantity 解析容量, 如 10Gi / 500Mi
func parseStorageQuantity(value string) (resource.Quantity, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return resource.Quantity{}, errors.New("容量不能为空")
	}
	qty, err := resource.ParseQuantity(value)
	if err != nil {
		return resource.Quantity{}, fmt.Errorf("容量格式不正确: %s, 例如 10Gi", value)
	}
	if qty.Sign() <= 0 {
		return resource.Quantity{}, errors.New("容量必须大于 0")
	}
	return qty, nil
}

// parseAccessModes 校验访问模式, 至少选择一个
func parseAccessModes(values []string) ([]corev1.PersistentVolumeAccessMode, error) {
	modes := make([]corev1.PersistentVolumeAccessMode, 0, len(values))
	seen := map[string]bool{}
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			continue
		}
		if !validAccessModes[v] {
			return nil, fmt.Errorf("不支持的访问模式: %s", v)
		}
		seen[v] = true
		modes = append(modes, corev1.PersistentVolumeAccessMode(v))
	}
	if len(modes) == 0 {
		return nil, errors.New("请至少选择一种访问模式")
	}
	return modes, nil
}

// parseVolumeMode 卷模式, 默认 Filesystem
func parseVolumeMode(value string) (corev1.PersistentVolumeMode, error) {
	switch mode := corev1.PersistentVolumeMode(strings.TrimSpace(value)); mode {
	case "":
		return corev1.PersistentVolumeFilesystem, nil
	case corev1.PersistentVolumeFilesystem, corev1.PersistentVolumeBlock:
		return mode, nil
	default:
		return "", fmt.Errorf("不支持的卷模式: %s, 可选 Filesystem / Block", value)
	}
}

// cleanStrings 去掉空白项
func cleanStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			result = append(result, v)
		}
	}
	return result
}

// cleanStringMap 去掉空键, 空 map 返回 nil
func cleanStringMap(values map[string]string) map[string]string {
	result := make(map[string]string, len(values))
	for k, v := range values {
		if k = strings.TrimSpace(k); k != "" {
			result[k] = v
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}

// withAliasDescribe 合并注解并写入别名 / 描述
func withAliasDescribe(annotations map[string]string, alias, describe string) map[string]string {
	result := map[string]string{}
	for k, v := range annotations {
		result[k] = v
	}
	setOrDelete(result, "alias", strings.TrimSpace(alias))
	setOrDelete(result, "describe", strings.TrimSpace(describe))
	if len(result) == 0 {
		return nil
	}
	return result
}
