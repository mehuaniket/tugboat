# Tugboat

Minimal version of `openkruise/kruise`. It was named Tugboat because it does minimal tasks with a small footprint compared to Kruise.
## Description
Introducing a lightweight Kubernetes operator for BroadcastJobs, enabling efficient pod deployment across your cluster.

### Key Features:

- **Effortless Node-Wide Deployments:** Seamlessly create BroadcastJobs to launch pods on all eligible nodes, streamlining cluster-wide operations.
- **Precise Node Selection:** Customize pod placement with node selectors for granular control.
- **Restart Management:** Define restart limits for individual pods to ensure resilience and manage potential failures.
- **Minimal Footprint:** Designed for efficiency with a smaller resource footprint than OpenKruise, ideal for resource-conscious environments.
- **Perfect for Learning Go:** Built as a learning project, it showcases Go development practices while creating a valuable Kubernetes tool.
Ideal for:

- Cluster-wide maintenance tasks
- Diagnostics and data collection
- Distributing workloads across all nodes
- Get started today and experience effortless cluster-wide pod deployments!

## Install 

```bash
helm repo add tugboat https://mehuaniket.github.io/tools/tugboat/index.yaml
helm install tugboat tugboat/tugboat
```

## Uninstall

- Create cronbroadcast job. `touch cronbroadcast.yaml`

```yaml
apiVersion: apps.tugboat.cloudrasayan.com/v1
kind: CronBroadcastJob
metadata:
  labels:
    app.kubernetes.io/name: cronbroadcastjob
    app.kubernetes.io/instance: cronbroadcastjob-sample
    app.kubernetes.io/part-of: tugboat
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: tugboat
  name: cronbroadcastjob-sample
spec:
  schedule: "*/2 * * * *"
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  broadcastJobTemplate:
    template:
      spec:
        containers:
          - name: helloworld
            image: hello-world
    # restartLimit: 3 #number of retry allowed for each node
    labels:
      kubernetes.io/hostname: microk8s-vm
    nodeSelector:
      kubernetes.io/os: linux
    cleanupafter: 500s
```

- Apply

```
kubectl apply -f cronbroadcast.yaml
```


## License

Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

