[![CI](https://github.com/jkroepke/aad-group-msteams-tag-sync/workflows/CI/badge.svg)](https://github.com/jkroepke/aad-group-msteams-tag-sync/actions?query=workflow%3ACI)
[![GitHub license](https://img.shields.io/github/license/jkroepke/aad-group-msteams-tag-sync)](https://github.com/jkroepke/aad-group-msteams-tag-sync/blob/master/LICENSE.txt)
[![Current Release](https://img.shields.io/github/release/jkroepke/aad-group-msteams-tag-sync.svg)](https://github.com/jkroepke/aad-group-msteams-tag-sync/releases/latest)
[![GitHub all releases](https://img.shields.io/github/downloads/jkroepke/aad-group-msteams-tag-sync/total?logo=github)](https://github.com/jkroepke/aad-group-msteams-tag-sync/releases/latest)
[![codecov](https://codecov.io/gh/jkroepke/aad-group-msteams-tag-sync/graph/badge.svg?token=66VT000UYO)](https://codecov.io/gh/jkroepke/aad-group-msteams-tag-sync)

# aad-group-msteams-tag-sync

Utility to sync AAD Groups with MS Teams Tags using Microsoft Graph API.

At this moment, MS Teams Tags only support up to 25 members and require at least 1 member.

# Requirements

Service Principal or equivalent with the following permissions:

* `GroupMember.Read.All`
* `Team.ReadBasic.All`
* `TeamworkTag.ReadWrite.All`

If you are using a Service Principal, ensure that you grant `Application`. `Delegated` permission won't work.

## Authentication

aad-group-msteams-tag-sync supports all authentication supported by Azure SDK for Go.
You have to declare one set of the environment variables below.

### Service principal with a secret

| Variable name         | Value                                        |
|-----------------------|----------------------------------------------|
| `AZURE_CLIENT_ID`     | Application ID of an Azure service principal |
| `AZURE_TENANT_ID`     | ID of the application's Azure AD tenant      |
| `AZURE_CLIENT_SECRET` | Password of the Azure service principal      |

### Service principal with certificate

| Variable name                   | Value                                                                          |
|---------------------------------|--------------------------------------------------------------------------------|
| `AZURE_CLIENT_ID`               | Application ID of an Azure service principal                                   |
| `AZURE_TENANT_ID`               | ID of the application's Azure AD tenant                                        |
| `AZURE_CLIENT_CERTIFICATE_PATH` | Path to a certificate file including private key (without password protection) |

### Use a managed identity

| Variable name     | Value                                                                              |
|-------------------|------------------------------------------------------------------------------------|
| `AZURE_CLIENT_ID` | User-assigned managed client id. Can be avoid, if a system assign identity is used |
| `AZURE_TENANT_ID` | ID of the application's Azure AD tenant                                            |

### Supporting documentation

- [graph API reference](https://docs.microsoft.com/en-us/graph/api/overview?view=graph-rest-1.0) (includes required permissions)

# Usage

```
Usage of ./aad-group-msteams-tag-sync:
  -config string
        path to config file
```

See [config.example.yaml](./config.example.yaml) for how to set up a sync config file.
`aad-group-msteams-tag-sync` terminate after successful sync.
