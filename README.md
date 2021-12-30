# prom-alert-webhook
----
用于prometheus使用alertmanager的webhook。
目前支持：
- 短信
    - 容联云
    - 阿里云
- 钉钉
- 企业微信
## 部署
1、下载代码
```shell script
git clone https://github.com/landyli/prom-alert-webhook.git
```
2、编译
```shell script
sh build.sh
```

3、打包镜像
```shell script
docker build -f Dockerfile.build -t registry.cn-hangzhou.aliyuncs.com/liyan/prom-alert-webhook:v0.0.6 .
```
注：镜像地址更换成自己的仓库地址  
4、推送镜像到镜像仓库
```shell script
docker push registry.cn-hangzhou.aliyuncs.com/liyan/prom-alert-webhook:v0.0.6
```
5、修改项目目录下的prom-alert-webhook.yaml
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: alert-webhook-conf
  namespace: monitoring
data:
  conf.yaml: |
      adapter:
          - sms
          - wechat
          - dingTalk
    
      # 短信平台
      sms:
        enable: false
        adapter_name: "RongLianYun"
    
        # 容联云短信平台
        RongLianYun:
          baseUrl: "https://app.cloopen.com:8883"
          accountSid: "xxxxx"
          appToken: "xxxxxx"
          appId: "xxxx"
          templateId: "xxxx"
          phones: ["1811111111"]
    
        # 阿里云短信平台
        AliYun:
          aliRegion: "cn-hangzhou"
          accessKeyId: "xxxx"
          accessSecret: "xxxx"
          phoneNumbers: "11111111111,22222222222"
          signName: "xxxx"
          templateCode: "xxxx"
      wechat:
        enable: false
        toUser: "Joker|Jase"
        agentId: "1000002"
        corpId: "xxxxxx"
        corpSecret: "xxxxxx"
    
      dingTalk:
        enable: true
        secret: "SEC94xxxxxx"
        access_token: "b438d4ce40c0xxxxx" 
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: prom-alert-webhook
  namespace: monitoring
spec:
  replicas: 1
  selector:
    matchLabels:
      app: prom-alert-webhook
  template:
    metadata:
      labels:
        app: prom-alert-webhook
    spec:
      containers:
        - name: prom-alert-webhook
          image: registry.cn-hangzhou.aliyuncs.com/liyan/prom-alert-webhook:v0.0.6
          imagePullPolicy: IfNotPresent
          livenessProbe:
            failureThreshold: 3
            httpGet:
              path: /healthCheck
              port: 9000
              scheme: HTTP
            initialDelaySeconds: 30
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 2
          readinessProbe:
              failureThreshold: 3
              httpGet:
                path: /healthCheck
                port: 9000
                scheme: HTTP
              initialDelaySeconds: 30
              periodSeconds: 10
              successThreshold: 1
              timeoutSeconds: 2
          ports:
            - name: app-port
              containerPort: 9000
              protocol: TCP
          resources:
            limits:
              cpu: 500m
              memory: 1Gi
            requests:
              cpu: 100m
              memory: 500Mi
          volumeMounts:
            - name: alert-webhook-conf
              mountPath: /app/conf/conf.yaml
              subPath: conf.yaml
      volumes:
        - name: alert-webhook-conf
          configMap:
            name: alert-webhook-conf
            defaultMode: 420
        - name: localtime
          hostPath:
            path: /etc/localtime
---
apiVersion: v1
kind: Service
metadata:
  name: prom-alert-webhook
  namespace: monitoring
spec:
  selector:
    app: prom-alert-webhook
  ports:
    - name: app-port
      port: 9000
      targetPort: 9000
      protocol: TCP
```
到自己购买的短信服务获取对应的信息。  
7、部署yaml文件
```shell script
kubectl apply -f prom-alert-webhook.yaml
```
8、修改alertmanager的报警媒介
```shell script
 ......
      - receiver: sms 
        group_wait: 10s
        match:
          filesystem: node
    receivers:
    - name: 'sms'
      webhook_configs:
      - url: "http://prom-alert-webhook.svc:9000"
        send_resolved: true
......
```

9、模板示例
```shell script

{{ define "wechat.default.message" }}
{{- if gt (len .Alerts.Firing) 0 -}}
{{- range $index, $alert := .Alerts -}}
{{- if eq $index 0 }}
==========异常告警==========
告警类型: {{ $alert.Labels.alertname }}
告警级别: {{ $alert.Labels.severity }}
告警详情: {{ $alert.Annotations.message }}{{ $alert.Annotations.description}};{{$alert.Annotations.summary}}
故障时间: {{ ($alert.StartsAt.Add 28800e9).Format "2006-01-02 15:04:05" }}
{{- if gt (len $alert.Labels.instance) 0 }}
实例信息: {{ $alert.Labels.instance }}
{{- end }}
{{- if gt (len $alert.Labels.namespace) 0 }}
命名空间: {{ $alert.Labels.namespace }}
{{- end }}
{{- if gt (len $alert.Labels.node) 0 }}
节点信息: {{ $alert.Labels.node }}
{{- end }}
{{- if gt (len $alert.Labels.pod) 0 }}
实例名称: {{ $alert.Labels.pod }}
{{- end }}
============END============
{{- end }}
{{- end }}
{{- end }}
{{- if gt (len .Alerts.Resolved) 0 -}}
{{- range $index, $alert := .Alerts -}}
{{- if eq $index 0 }}
==========异常恢复==========
告警类型: {{ $alert.Labels.alertname }}
告警级别: {{ $alert.Labels.severity }}
告警详情: {{ $alert.Annotations.message }}{{ $alert.Annotations.description}};{{$alert.Annotations.summary}}
故障时间: {{ ($alert.StartsAt.Add 28800e9).Format "2006-01-02 15:04:05" }}
恢复时间: {{ ($alert.EndsAt.Add 28800e9).Format "2006-01-02 15:04:05" }}
{{- if gt (len $alert.Labels.instance) 0 }}
实例信息: {{ $alert.Labels.instance }}
{{- end }}
{{- if gt (len $alert.Labels.namespace) 0 }}
命名空间: {{ $alert.Labels.namespace }}
{{- end }}
{{- if gt (len $alert.Labels.node) 0 }}
节点信息: {{ $alert.Labels.node }}
{{- end }}
{{- if gt (len $alert.Labels.pod) 0 }}
实例名称: {{ $alert.Labels.pod }}
{{- end }}
============END============
{{- end }}
{{- end }}
{{- end }}
{{- end }}
```
