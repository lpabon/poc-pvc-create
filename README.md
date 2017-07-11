# Demo of PVC control

```
$ kubectl --kubeconfig=kubeconfig get storageclass
NAME                 TYPE
gluster.qm.gluster   kubernetes.io/glusterfs 

$./pvc-create -kubeconfig=kubeconfig -storageclass gluster.qm.gluster
fatal: No names found, cannot describe anything.
go build -ldflags "-extldflags '-z relro -z now'" -o pvc-create main.go
pvc-create INFO 2017/07/11 00:33:54 main.go:64: Starting
pvc-create INFO 2017/07/11 00:33:55 main.go:77: connection to Kubernetes established. Cluster version v1.6.6+coreos.1
pvc-create INFO 2017/07/11 00:33:55 main.go:152: Number of PVCs: 0
pvc-create INFO 2017/07/11 00:33:55 main.go:194: PVC pvc-create-sample-pvc submitted
pvc-create INFO 2017/07/11 00:33:55 main.go:202: Waiting for PVC to be bound
pvc-create INFO 2017/07/11 00:34:07 main.go:152: Number of PVCs: 1
pvc-create INFO 2017/07/11 00:34:07 main.go:159: PVC: pvc-create-sample-pvc
        Volume name: pvc-2643f31f-65f2-11e7-aaf2-022ac9c4110c
        Size Available: 10
pvc-create INFO 2017/07/11 00:34:07 main.go:230: Deleting PVC pvc-create-sample-pvc
pvc-create INFO 2017/07/11 00:34:11 main.go:152: Number of PVCs: 0

```