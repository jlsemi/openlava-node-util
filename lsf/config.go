package lsf

const bqueueConfig = `
{{ range . }}
Begin Queue
QUEUE_NAME   = {{ .QueueName }}
PRIORITY     = 30
NICE         = 20
USERS        = {{ .Users }}
HOSTS = {{ range .Hosts }} {{ . }} {{ end }}
End Queue
{{ end }}
`

const clusterConfig = `
Begin   ClusterAdmins
Administrators = lsfadmin
End    ClusterAdmins


Begin   Host
HOSTNAME          model          type  server  r1m  RESOURCES
{{ range . }}{{ .HostName }}  IntelI5  linux {{ .HostType }} 3.5 (cs)
{{ end }}
End Host 

Begin ResourceMap
RESOURCENAME  LOCATION
End ResourceMap
`

const bhostConfig = `
Begin Host
HOST_NAME     MXJ JL/U   r1m    pg    ls     tmp  DISPATCH_WINDOW  # Keywords
{{ range . }}{{ .HostName }} {{ .MaxNodes }}  ()  ()  ()  ()  ()  ()
{{ end }}
End Host
`
