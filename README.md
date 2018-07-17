_This is a work in progress_

vaultsmith
==========

_noun_

One who provisions Vaults


🤨

It's a play on locksmith.

😑

What
----

vaultsmith is a tool for provisioning instances of Hashicorp Vault, for 
example by setting up roles and access permissions. The goal is to support a 
_declarative_ style of configuration, i.e. the configuration reflects the 
final state, and executions are idempotent.


This project is under active development, and third party contributions are 
most welcome. 

Documentation is lacking at this point, and obviously given its early stage 
it should only be used on test servers.

Concept
-------
A couple of years ago, Hashicorp published a blog post 
["Codifying Vault Policies and Configuration"](https://www.hashicorp.com/blog/codifying-vault-policies-and-configuration.html). 
We used a heavily modified version of their scripts to get us going with Vault. Vaultsmith is 
really just an extension of that idea, except that it uses the official Go client, and
obviously is written in Go as well.

Essentially, the directory structure (document-path) is mapped to the API endpoints of Vault,
and the contents of the document within is posted to Vault, using the built-in Vault client. 
It gets more complicated when you consider endpoints such as sys/auth and sys/policy have
special methods in the Vault client, so these directories are assigned specific handlers which
call the appropriate methods.

Installation
--------
#### Native Go
```bash
go get github.com/starlingbank/vaultsmith
```

#### Docker
The image is not published, but you can build it after go get with:
```bash
cd $GOPATH/src/github.com/starlingbank/vaultsmith
make docker
```

Usage
-----

```
vaultsmith -h
Usage of vaultsmith:
  --document-path string
    	The root directory of the configuration. Can be a local directory, local gz tarball or http url to a gz tarball. (default "./example")
  --role string
    	The Vault role to authenticate as
  --template-file string
    	JSON file containing template mappings. If not specified, vaultsmith will look for "template.json" in the base of the document path.

Vault authentication is handled by environment variables (the same ones as the Vault client, as vaultsmith uses the same code). So ensure VAULT_ADDR and VAULT_TOKEN are set.
```

Examples
--------
Run up a test vault server and export your token:
```bash
docker run -p 8200:8200 vault:latest
export VAULT_TOKEN=$root_token
export VAULT_ADDR=http://localhost:8200
```
Run vaultsmith and it should apply the example document set:
```bash
vaultsmith -document-path https://raw.githubusercontent.com/starlingbank/vaultsmith/master/example/example.tgz
```
