# Derivatives featureset

Derivatives that inherit directly `system/cos` have the following defaults

<!-- TOC -->

- [Derivatives featureset](#derivatives-featureset)
    - [Login](#login)
    - [Install](#install)
    - [Upgrades](#upgrades)
    - [Reset state](#reset-state)
        - [Recovery partition](#recovery-partition)
        - [From ISO](#from-iso)
    - [File system layout](#file-system-layout)
    - [Persistent changes](#persistent-changes)
        - [Available stages](#available-stages)
            - [initramfs](#initramfs)
            - [boot](#boot)
            - [fs](#fs)
            - [network](#network)
            - [reconcile](#reconcile)
    - [cOS runtime features](#cos-runtime-features)
    - [OEM customizations](#oem-customizations)
        - [Default OEM](#default-oem)
    - [Configuration reference](#configuration-reference)
        - [Compatibility with Cloud Init format](#compatibility-with-cloud-init-format)
        - [stages.<stageID>.[<stepN>].name](#stagesstageidstepnname)
        - [stages.<stageID>.[<stepN>].files](#stagesstageidstepnfiles)
        - [stages.<stageID>.[<stepN>].directories](#stagesstageidstepndirectories)
        - [stages.<stageID>.[<stepN>].dns](#stagesstageidstepndns)
        - [stages.<stageID>.[<stepN>].hostname](#stagesstageidstepnhostname)
        - [stages.<stageID>.[<stepN>].sysctl](#stagesstageidstepnsysctl)
        - [stages.<stageID>.[<stepN>].authorized_keys](#stagesstageidstepnauthorized_keys)
        - [stages.<stageID>.[<stepN>].node](#stagesstageidstepnnode)
        - [stages.<stageID>.[<stepN>].users](#stagesstageidstepnusers)
        - [stages.<stageID>.[<stepN>].ensure_entities](#stagesstageidstepnensure_entities)
        - [stages.<stageID>.[<stepN>].delete_entities](#stagesstageidstepndelete_entities)
        - [stages.<stageID>.[<stepN>].modules](#stagesstageidstepnmodules)
        - [stages.<stageID>.[<stepN>].systemctl](#stagesstageidstepnsystemctl)
        - [stages.<stageID>.[<stepN>].environment](#stagesstageidstepnenvironment)
        - [stages.<stageID>.[<stepN>].environment_file](#stagesstageidstepnenvironment_file)
        - [stages.<stageID>.[<stepN>].timesyncd](#stagesstageidstepntimesyncd)
        - [stages.<stageID>.[<stepN>].commands](#stagesstageidstepncommands)
        - [stages.<stageID>.[<stepN>].datasource](#stagesstageidstepndatasource)

<!-- /TOC -->

## Login

You can login with the user `root` and `cos`. That's a live ISO and no changes will be persisted. You can change this by overriding `/system/oem/04_accounting.yaml` in the derivative spec file. ([Example](https://github.com/rancher-sandbox/cos-toolkit-sample-repo/blob/00c0b4abf8225224c1c177f5b3bd818c7b091eaf/packages/sampleOS/build.yaml#L13))

## Install

To install run `cos-installer <device>` to start the installation process. Remove the ISO and reboot.

_Note_: `cos-installer` supports other options as well. Run `cos-installer --help` to see a complete help.

## Upgrades

To upgrade the system, just run `cos-upgrade` and reboot.

cOS during installation sets two `.img` images files in the `COS_STATE` partition:
- `/cOS/active.img` labeled `COS_ACTIVE`: Where `cOS` typically boots from
- `/cOS/passive.img` labeled `COS_PASSIVE`: Where `cOS` boots for fallback

Those are used by the upgrade mechanism to prepare and install a pristine `cOS` each time an upgrade is attempted.

To specify a single docker image to upgrade to  instead of the regular upgrade channels, run `cos-upgrade --docker-image image`.

_Note_ by default `cos-upgrade --docker-image` checks images to the notary registry server for valid signatures for the images tag. To disable image verification, run `cos-upgrade --no-verify --docker-image`.

See the [sample repository](https://github.com/rancher-sandbox/cos-toolkit-sample-repo#system-upgrades) readme on how to tweak the upgrade channels for the derivative.

## Reset state

### Recovery partition

cOS can be recovered anytime from the `cOS recovery` partition by running `cos-reset`. This will regenerate the bootloader and the images in `COS_STATE` by using the recovery image created during installation.

The recovery partition can also be upgraded by running `cos-upgrade --recovery` in the standard partitions used for boot.

### From ISO
The ISO can be also used as a recovery medium: type `cos-upgrade` from a LiveCD. It will then try to upgrade the image of the active partition installed in the system.

## File system layout

As cOS is an immutable distribution, the file system layout is a core aspect. A running `cOS` derivative will look as follows:

```
/usr/local - persistent (COS_PERSISTENT)
/oem - persistent (COS_OEM)
/etc - ephemeral
/usr - read only
/ immutable
```

Any changes that are not specified by cloud-init are not persisting across reboots.

## Persistent changes

By default cOS reads and executes cloud-init files in (lexicopgrahic) sequence present in `/usr/local/cloud-config` and `/oem` during boot. It is also possible to run cloud-init file in a different location from boot cmdline by using  the `cos.setup=..` option.

For example, if you want to change `/etc/issue` of the system persistently, you can create `/usr/local/cloud-config/90_after_install.yaml` with the following content:

```yaml
# The following is executed before fs is setted up:
stages:
    fs:
        - name: "After install"
          files:
          - path: /etc/issue
            content: |
                    Welcome, have fun!
            permissions: 0644
            owner: 0
            group: 0
          systemctl:
            disable:
            - wicked
        - name: "After install (second step)"
          files:
          - path: /etc/motd
            content: |
                    Welcome, have more fun!
            permissions: 0644
            owner: 0
            group: 0
```

For more examples, `/system/oem` contains files used to configure on boot a pristine `cOS`. Mind to not edit those directly, but copy them or apply local changes to `/usr/local/cloud-config`. See the OEM section below.

### Available stages

Cloud-init files are applied in 5 different phases: `boot`, `network`, `fs`, `initramfs` and `reconcile`. All the available cloud-init keywords can be used in each stage. Additionally, it's possible also to hook before or after a stage has run, each one has a specific stage which is possible to run steps: `boot.after`, `network.before`, `fs.after` etc.

#### initramfs

This is the earliest stage, running before switching root. Here you can apply radical changes to the booting setup of `cOS`.

#### boot

This stage is executed after initramfs has switched root, during the `systemd` bootup process.

#### fs

This stage is executed when fs is mounted and is guaranteed to have access to `COS_STATE` and `COS_PERSISTENT`.

#### network

This stage is executed when network is available

#### reconcile

This stage is executed `5m` after boot and periodically each `60m`.

## cOS runtime features

cOS ships default cloud-init configurations files that are available under `/system/features` for example purposes, and to quickly enable testing features.

Features can be enabled/disabled with `cos-feature`. For example, after install, to enable `k3s` it's sufficient to type `cos-feature enable k3s` and reboot.

See `cos-feature list` for the available features.


```
$> cos-feature list

====================
cOS features list

To enable, run: cos-feature enable <feature>
To disable, run: cos-feature disable <feature>
====================

- carrier
- harvester
- k3s
- vagrant (enabled)
...
```

You are encouraged to copy them over to `/usr/local/cloud-config` or `/oem` and customize them as you see fit.

## OEM customizations

It is possible to install a custom cloud-init file during install with `--config` to `cos-installer` or, it's possible to add more files manually to the `/oem` folder after installation.

### Default OEM

By default, `cOS` ships a set of default configurations which can be found under `/system/oem`. This is to setup e.g. the default root password and the upgrade channel. 

```
/system/oem/00_rootfs.yaml - defines the rootfs mountpoint layout setting
/system/oem/01_defaults.yaml - systemd defaults (keyboard layout, timezone)
/system/oem/02_upgrades.yaml - Settings for channel upgrades
/system/oem/03_branding.yaml - Branding setting, Derivative name, /etc/issue content
/system/oem/04_accounting.yaml - Default user/pass
/system/oem/05_network.yaml - Default network setup
/system/oem/06_recovery.yaml - Executes additional commands when booting in recovery mode
```

If you are building a cOS derivative, and plan to release upgrades, you must override (or create a new file under `/system/oem`) the `/system/oem/02_upgrades.yaml` pointing to the docker registry used to deliver upgrades.



## Configuration reference

Below is a reference of all keys available in the cloud-init style files.

```yaml
stages:
   # "network" is the stage
   network:
     - files:
        - path: /tmp/foo
          content: |
                    test
          permissions: 0777
          owner: 1000
          group: 100
       commands:
        - echo "test"
       modules:
       - nvidia
       environment:
         FOO: "bar"
       systctl:
         debug.exception-trace: "0"
       hostname: "foo"
       systemctl:
         enable:
         - foo
         disable:
         - bar
         start:
         - baz
         mask:
         - foobar
       authorized_keys:
          user:
          - "github:mudler"
          - "ssh-rsa ...."
       dns:
         path: /etc/resolv.conf
         nameservers:
         - 8.8.8.8
       ensure_entities:
       -  path: /etc/passwd
          entity: |
                  kind: "user"
                  username: "foo"
                  password: "pass"
                  uid: 0
                  gid: 0
                  info: "Foo!"
                  homedir: "/home/foo"
                  shell: "/bin/bash"
       delete_entities:
       -  path: /etc/passwd
          entity: |
                  kind: "user"
                  username: "foo"
                  password: "pass"
                  uid: 0
                  gid: 0
                  info: "Foo!"
                  homedir: "/home/foo"
                  shell: "/bin/bash"
       datasource:
         providers:
         - "aws"
         - "digitalocean"
         path: "/etc/cloud-data"
```


### Compatibility with Cloud Init format

A subset of the official [cloud-config spec](http://cloudinit.readthedocs.org/en/latest/topics/format.html#cloud-config-data) is implemented by yip. 

If a yaml file starts with `#cloud-config` it is parsed as a standard cloud-init, associated it to the yip `boot` stage. For example:

```yaml
#cloud-config
users:
- name: "bar"
  passwd: "foo"
  groups: "users"
  ssh_authorized_keys:
  - faaapploo
ssh_authorized_keys:
  - asdd
runcmd:
- foo
hostname: "bar"
write_files:
- encoding: b64
  content: CiMgVGhpcyBmaWxlIGNvbnRyb2xzIHRoZSBzdGF0ZSBvZiBTRUxpbnV4
  path: /foo/bar
  permissions: "0644"
  owner: "bar"
```

### `stages.<stageID>.[<stepN>].name`

A description of the stage step. Used only when printing output to console.

### `stages.<stageID>.[<stepN>].files`

A list of files to write to disk.

```yaml
stages:
   default:
     - files:
        - path: /tmp/bar
          content: |
                    #!/bin/sh
                    echo "test"
          permissions: 0777
          owner: 1000
          group: 100
```

### `stages.<stageID>.[<stepN>].directories`

A list of directories to be created on disk. Runs before `files`.

```yaml
stages:
   default:
     - name: "Setup folders"
       directories:
       - path: "/etc/foo"
         permissions: 0600
         owner: 0
         group: 0
```

### `stages.<stageID>.[<stepN>].dns`

A way to configure the `/etc/resolv.conf` file.

```yaml
stages:
   default:
     - name: "Setup dns"
       dns:
         nameservers:
         - 8.8.8.8
         - 1.1.1.1
         search:
         - foo.bar
         options:
         - ..
         path: "/etc/resolv.conf.bak"
```
### `stages.<stageID>.[<stepN>].hostname`

A string representing the machine hostname. It sets it in the running system, updates `/etc/hostname` and adds the new hostname to `/etc/hosts`.

```yaml
stages:
   default:
     - name: "Setup hostname"
       hostname: "foo"
```
### `stages.<stageID>.[<stepN>].sysctl`

Kernel configuration. It sets `/proc/sys/<key>` accordingly, similarly to `sysctl`.

```yaml
stages:
   default:
     - name: "Setup exception trace"
       systctl:
         debug.exception-trace: "0"
```

### `stages.<stageID>.[<stepN>].authorized_keys`

A list of SSH authorized keys that should be added for each user.
SSH keys can be obtained from GitHub user accounts by using the format github:${USERNAME},  similarly for Gitlab with gitlab:${USERNAME}.

```yaml
stages:
   default:
     - name: "Setup exception trace"
       authorized_keys:
         mudler:
         - github:mudler
         - ssh-rsa: ...
```

### `stages.<stageID>.[<stepN>].node`

If defined, the node hostname where this stage has to run, otherwise it skips the execution. The node can be also a regexp in the Golang format.

```yaml
stages:
   default:
     - name: "Setup logging"
       node: "bastion"
```

### `stages.<stageID>.[<stepN>].users`

A map of users and password to set. Passwords can be also encrypted.

```yaml
stages:
   default:
     - name: "Setup users"
       users:
          bastion: "strongpassword"
```

### `stages.<stageID>.[<stepN>].ensure_entities`

A `user` or a `group` in the [entity](https://github.com/mudler/entities) format to be configured in the system

```yaml
stages:
   default:
     - name: "Setup users"
       ensure_entities:
       -  path: /etc/passwd
          entity: |
                  kind: "user"
                  username: "foo"
                  password: "x"
                  uid: 0
                  gid: 0
                  info: "Foo!"
                  homedir: "/home/foo"
                  shell: "/bin/bash"
```
### `stages.<stageID>.[<stepN>].delete_entities`

A `user` or a `group` in the [entity](https://github.com/mudler/entities) format to be pruned from the system

```yaml
stages:
   default:
     - name: "Setup users"
       delete_entities:
       -  path: /etc/passwd
          entity: |
                  kind: "user"
                  username: "foo"
                  password: "x"
                  uid: 0
                  gid: 0
                  info: "Foo!"
                  homedir: "/home/foo"
                  shell: "/bin/bash"
```
### `stages.<stageID>.[<stepN>].modules`

A list of kernel modules to load.

```yaml
stages:
   default:
     - name: "Setup users"
       modules:
       - nvidia
```
### `stages.<stageID>.[<stepN>].systemctl`

A list of systemd services to `enable`, `disable`, `mask` or `start`.

```yaml
stages:
   default:
     - name: "Setup users"
       systemctl:
         enable:
          - systemd-timesyncd
          - cronie
         mask:
          - purge-kernels
         disable:
          - crond
         start:
          - cronie
```
### `stages.<stageID>.[<stepN>].environment`

A map of variables to write in `/etc/environment`, or otherwise specified in `environment_file`

```yaml
stages:
   default:
     - name: "Setup users"
       environment:
         FOO: "bar"
```
### `stages.<stageID>.[<stepN>].environment_file`

A string to specify where to set the environment file

```yaml
stages:
   default:
     - name: "Setup users"
       environment_file: "/home/user/.envrc"
       environment:
         FOO: "bar"
```
### `stages.<stageID>.[<stepN>].timesyncd`

Sets the `systemd-timesyncd` daemon file (`/etc/system/timesyncd.conf`) file accordingly. The documentation for `timesyncd` and all the options can be found [here](https://www.freedesktop.org/software/systemd/man/timesyncd.conf.html).

```yaml
stages:
   default:
     - name: "Setup NTP"
       systemctl:
         enable:
         - systemd-timesyncd
       timesyncd:
          NTP: "0.pool.org foo.pool.org"
          FallbackNTP: ""
          ...
```

### `stages.<stageID>.[<stepN>].commands`

A list of arbitrary commands to run after file writes and directory creation.

```yaml
stages:
   default:
     - name: "Setup something"
       commands:
         - echo 1 > /bar
```

### `stages.<stageID>.[<stepN>].datasource`

Sets to fetch user data from the specified cloud providers. It populates
provider specific data into `/run/config` folder and the custom user data
is stored into the provided path.


```yaml
stages:
   default:
     - name: "Fetch cloud provider's user data"
       datasource:
         providers:
         - "aws"
         - "digitalocean"
         path: "/etc/cloud-data"
```
