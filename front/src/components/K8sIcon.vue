<template>
  <img
    :src="iconSrc"
    :alt="alt || name"
    class="k8s-icon"
    :style="iconStyle"
    draggable="false"
  />
</template>

<script setup lang="ts">
import { computed } from 'vue'
import k8sIcon from '@/assets/logo-k8s.svg'
import groupIcon from '@/assets/icons/k8s/group.svg'
import nsIcon from '@/assets/icons/k8s/ns.svg'
import nodeIcon from '@/assets/icons/k8s/node.svg'
import podIcon from '@/assets/icons/k8s/pod.svg'
import rsIcon from '@/assets/icons/k8s/rs.svg'
import deployIcon from '@/assets/icons/k8s/deploy.svg'
import dsIcon from '@/assets/icons/k8s/ds.svg'
import stsIcon from '@/assets/icons/k8s/sts.svg'
import netpolIcon from '@/assets/icons/k8s/netpol.svg'
import svcIcon from '@/assets/icons/k8s/svc.svg'
import ingIcon from '@/assets/icons/k8s/ing.svg'
import epIcon from '@/assets/icons/k8s/ep.svg'
import cmIcon from '@/assets/icons/k8s/cm.svg'
import secretIcon from '@/assets/icons/k8s/secret.svg'
import saIcon from '@/assets/icons/k8s/sa.svg'
import jobIcon from '@/assets/icons/k8s/job.svg'
import cronjobIcon from '@/assets/icons/k8s/cronjob.svg'
import volIcon from '@/assets/icons/k8s/vol.svg'
import pvIcon from '@/assets/icons/k8s/pv.svg'
import pvcIcon from '@/assets/icons/k8s/pvc.svg'
import scIcon from '@/assets/icons/k8s/sc.svg'
import crdIcon from '@/assets/icons/k8s/crd.svg'
import etcdIcon from '@/assets/icons/k8s/etcd.svg'
import hpaIcon from '@/assets/icons/k8s/hpa.svg'
import quotaIcon from '@/assets/icons/k8s/quota.svg'
import limitsIcon from '@/assets/icons/k8s/limits.svg'

export type K8sIconName =
  | 'k8s'
  | 'group'
  | 'ns'
  | 'node'
  | 'pod'
  | 'rs'
  | 'deploy'
  | 'ds'
  | 'sts'
  | 'netpol'
  | 'svc'
  | 'ing'
  | 'ep'
  | 'cm'
  | 'secret'
  | 'sa'
  | 'job'
  | 'cronjob'
  | 'vol'
  | 'pv'
  | 'pvc'
  | 'sc'
  | 'crd'
  | 'etcd'
  | 'hpa'
  | 'quota'
  | 'limits'

const iconMap: Record<K8sIconName, string> = {
  k8s: k8sIcon,
  group: groupIcon,
  ns: nsIcon,
  node: nodeIcon,
  pod: podIcon,
  rs: rsIcon,
  deploy: deployIcon,
  ds: dsIcon,
  sts: stsIcon,
  netpol: netpolIcon,
  svc: svcIcon,
  ing: ingIcon,
  ep: epIcon,
  cm: cmIcon,
  secret: secretIcon,
  sa: saIcon,
  job: jobIcon,
  cronjob: cronjobIcon,
  vol: volIcon,
  pv: pvIcon,
  pvc: pvcIcon,
  sc: scIcon,
  crd: crdIcon,
  etcd: etcdIcon,
  hpa: hpaIcon,
  quota: quotaIcon,
  limits: limitsIcon
}

const props = withDefaults(
  defineProps<{
    name: K8sIconName
    size?: number | string
    alt?: string
  }>(),
  {
    size: 20,
    alt: ''
  }
)

const iconSrc = computed(() => iconMap[props.name] || k8sIcon)

const iconStyle = computed(() => {
  const s = typeof props.size === 'number' ? `${props.size}px` : props.size
  return {
    width: s,
    height: s
  }
})
</script>

<style scoped>
.k8s-icon {
  display: inline-block;
  flex-shrink: 0;
  object-fit: contain;
  vertical-align: middle;
  user-select: none;
}
</style>
