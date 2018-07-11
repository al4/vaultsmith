path_handlers
=============

Files in this package handle specific paths in the Vault configuration documents. For example, sys/auth needs to use a different API to sys/policy. 
Most paths will be simple document puts and should use the generic handler.

New handlers can be created by implementing the PathHandler interface.