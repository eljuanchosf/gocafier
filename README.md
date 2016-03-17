# gocafier

Este pequeño proyecto sirve para dar alivio a un rasgo de ansiedad que casi todos tenemos...
Cada vez que encargaba un envío por OCA, realmente me jodía no tener un servicio automatizado que me enviara notificaciones de acuerdo a los cambio de estado de los envíos. Tenía que ingresar al sitio web de OCA y consultarlo yo. Lo que se me ocurrió, viendo los requests y responses de ese sitio web, es que podría automatizar ese proceso y generar notificaciones al respecto.
Así nació **Gocafier**.

## Requisitos

* Golang (recomiendo usar [gvm](https://github.com/moovweb/gvm) para manejar versiones)
* [Godep](https://github.com/tools/godep)

## Instalación

> **Importante**: Estas instrucciones son para Linux y probablemente para Mac. No lo probé en Windows.

```
go get github.com/eljuanchosf/gocafier
godep go build
mv config.yml.example config.yml
```

## Configuración

### Servidor de correos

Con tu editor de texto favorito, abrí el archivo `config.yml` y configurá los datos de servidor de correos.

### Paquetes a buscar

Dentro de la key `packages` podes configurar un array de numeros de seguimiento.

Ejemplo:

```yaml
packages:
  - 00000000000000
  - 00000000000001
```

## Uso

**Gocafier** es fácil de usar.

### Ejecución

Para ver como usarlo, se puede ejecutar `./gocafier --help`:

```
$ ./gocafier --help
[2016-03-17 10:57:49.951451697 -0300 ART] Starting gocafier 1.0.0
usage: gocafier --smtp-user=SMTP-USER --smtp-pass=SMTP-PASS [<flags>]

Flags:
  --help               Show help (also see --help-long and --help-man).
  --debug              Enable debug mode. This disables emailing
  --cache-path=CACHE-PATH
                       Bolt Database path
  --ticker-time=3600s  Poller interval in secs
  --config-path="."    Set the Path to write profiling file
  --smtp-user=SMTP-USER
                       Sets the SMTP username
  --smtp-pass=SMTP-PASS
                       Sets the SMTP password
  --version            Show application version.
```

Lo más importante son los parámetros `--smtp-user` y `--smtp-pass`, en los que hay que especificar el usuario y contraseña del servidor de correo. Esos dos valores pueden también setearse mediante las variables de entorno `GOCAFIER_SMTP_USER` y `GOCAFIER_SMTP_PASSWORD`.

### Configurando el template

Podés configurar el template que Gocafier va a usar para enviar el email editando el archivo `email-template.html`. Se explica a sí mismo bastante bien.

## Contribuciones

Fork, branch, commit, PR. :)
