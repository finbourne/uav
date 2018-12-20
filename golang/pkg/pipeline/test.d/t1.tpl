{{- define "counter-fbn-prod" -}}
    {{ if . }}
      - get: counter-fbn-prod
        trigger: true
        passed: 
        {{ range .}}
        - {{ . }}
        {{ end }}
    {{ else }}
      - get: thing
    {{ end }}
{{- end -}}
