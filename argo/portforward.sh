 #!/bin/sh
 
 kubectl -n argo port-forward --address 0.0.0.0 service/argo-server 2746:2746