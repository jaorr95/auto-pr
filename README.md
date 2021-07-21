## AUTO-PR (Auto pull request)

Herramienta CLI escrita en GO 1.16.0 que permite realizar merge request y hacer el merge.

Tanto el merge request como el merge se realizan haciendo peticiones a la API de BITBUCKET.

El comando para construir el binario es:
```CGO_ENABLED=0 go build -tags netgo -a -v main.go```

Para visualizar los argumentos que se le puede pasar al cli se debe ejecutar:
```./auto-pr -h```