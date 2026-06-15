# CIVO Provider for Devsy

[![Open in Devsy!](https://img.shields.io/badge/open_in_devsy-8A2BE2?style=for-the-badge)](https://devsy.sh/open#https://github.com/devsy-org/devsy-provider-civo)

## Getting started

The provider is available for auto-installation using:

```sh
devsy provider add civo
devsy provider use civo
```

Follow the on-screen instructions to complete the setup.

Needed variables will be:

- CIVO_REGION
- CIVO_API_KEY

### Creating your first workspace with civo

After the initial setup, just use:

```sh
devsy up .
```

You'll need to wait for the machine and workspace setup.

### Customize the VM Instance

This provider has the following options:

| NAME               | REQUIRED | DESCRIPTION                             | DEFAULT                              |
|--------------------|----------|-----------------------------------------|--------------------------------------|
| CIVO_DISK_IMAGE    | false    | The disk image to use.                  | d927ad2f-5073-4ed6-b2eb-b8e61aef29a8 |
| CIVO_DISK_SIZE     | false    | The disk size to use.                   | 40                                   |
| CIVO_INSTANCE_TYPE | false    | The machine type to use.                | g3.large                             |
| CIVO_REGION        | true     | The civo cloud region to create the VM. |                                      |
| CIVO_API_KEY       | true     | The api key to use.                     |                                      |

Options can be set using `devsy provider set-options`, for example:

```sh
devsy provider set-options -o CIVO_REGION=LON1
```
