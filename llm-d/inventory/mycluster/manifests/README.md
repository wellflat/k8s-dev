# mycluster manifests

`inventory/mycluster/manifests` 配下は、Kubespray 管理外で追加した Kubernetes 関連 manifest をまとめる場所です。

## 構成

- `gpu/`
  - `nvidia-runtimeclass.yaml`: `RuntimeClass` `nvidia`
  - `nvidia-device-plugin.yaml`: NVIDIA device plugin DaemonSet
  - `gpu-smoke-test.yaml`: GPU 利用確認用の単発 Pod
- `llm-d/`
  - `agentgateway-values.yaml`: `agentgateway` Helm values
  - `llm-d-infra-values.yaml`: `llm-d-infra` Helm values
  - `vllm-hf-cache-pv.yaml`: vLLM cache 用 PV/PVC
  - `vllm-backend.yaml`: vLLM backend Deployment/Service
    - `llm-d-system` namespace の Secret `hf-token` から `HF_TOKEN` を注入する
  - `vllm-backend-route.yaml`: Gateway から vLLM backend への HTTPRoute

## 運用メモ

- `gpu/` 配下のうち、`gpu-smoke-test.yaml` は確認用です。常駐リソースではありません。
- `llm-d/` 配下は、Gateway 基盤と vLLM backend を再適用するときの作業用 manifest です。
- `vllm-backend` Service は `NodePort` `30080` で公開し、GPU ノード IP から直接アクセスできるようにしています。

## 再適用の順番

1. GPU ノード設定を Kubespray 側で反映する
2. GPU 関連 manifest を適用する
3. Gateway API / agentgateway / llm-d-infra を再適用する
4. vLLM backend と route を再適用する
5. 必要なら GPU と Gateway の疎通確認を行う

## 実行例

### 1. GPU ノード設定を Kubespray で反映

`host_vars/gpu-03.yml` の内容を反映するため、Kubespray を再実行します。

```bash
ansible-playbook -i inventory/mycluster/inventory.ini \
  --private-key ~/.ssh/id_ed25519_liquidinc \
  -b -K \
  cluster.yml
```

### 2. GPU 関連 manifest を適用

```bash
kubectl apply -f inventory/mycluster/manifests/gpu/nvidia-runtimeclass.yaml
kubectl apply -f inventory/mycluster/manifests/gpu/nvidia-device-plugin.yaml
```

確認用 Pod を使うときだけ追加で適用します。

```bash
kubectl apply -f inventory/mycluster/manifests/gpu/gpu-smoke-test.yaml
kubectl logs pod/gpu-smoke-test
kubectl delete -f inventory/mycluster/manifests/gpu/gpu-smoke-test.yaml
```

### 3. Gateway 基盤を再適用

必要な Helm repo が未登録なら先に追加します。

```bash
helm repo add agentgateway https://charts.agentgateway.dev
helm repo add llm-d https://llm-d.ai/helm-charts
helm repo update
```

`agentgateway` を再適用します。

```bash
helm upgrade --install agentgateway-crds agentgateway/agentgateway-crds \
  -n agentgateway-system \
  --create-namespace

helm upgrade --install agentgateway agentgateway/agentgateway \
  -n agentgateway-system \
  -f inventory/mycluster/manifests/llm-d/agentgateway-values.yaml
```

`llm-d-infra` を再適用します。

```bash
helm upgrade --install llm-d-infra llm-d/llm-d-infra \
  -n llm-d-system \
  --create-namespace \
  -f inventory/mycluster/manifests/llm-d/llm-d-infra-values.yaml
```

### 4. vLLM backend を再適用

先に Hugging Face token を Secret として作成または更新します。

```bash
kubectl -n llm-d-system create secret generic hf-token \
  --from-literal=HF_TOKEN='<your-huggingface-token>' \
  --dry-run=client -o yaml | kubectl apply -f -
```

```bash
kubectl apply -f inventory/mycluster/manifests/llm-d/vllm-hf-cache-pv.yaml
kubectl apply -f inventory/mycluster/manifests/llm-d/vllm-backend.yaml
kubectl apply -f inventory/mycluster/manifests/llm-d/vllm-backend-route.yaml
```

### 5. 動作確認

ノードと GPU device plugin を確認します。

```bash
kubectl get nodes -o wide
kubectl -n kube-system get pods -l k8s-app=nvidia-gpu-device-plugin -o wide
```

Gateway と vLLM backend を確認します。

```bash
kubectl -n llm-d-system get gateway,httproute,svc,pods
kubectl -n llm-d-system get pvc,pv
```

ローカル PC から GPU ノード IP に直接アクセスして vLLM backend を確認する例です。

```bash
curl -sS http://192.168.250.103:30080/health
curl -sS http://192.168.250.103:30080/v1/models
```

Python スクリプトで completion API を確認する例です。

```bash
export LLM_D_BASE_URL=http://192.168.250.103:30080/v1
export LLM_D_MODEL=gpt-oss-20b
uv run python llm-d/test_completion.py
```

Gateway 経由で OpenAI 互換 API を確認します。

```bash
kubectl run curl-models --rm -i --restart=Never \
  --image=curlimages/curl:8.12.1 \
  --command -- \
  curl -sS http://llm-d-infra-inference-gateway.llm-d-system.svc.cluster.local/v1/models
```
