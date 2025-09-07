 #!/bin/sh
 
 kubectl -n default port-forward --address 0.0.0.0 service/argo-workflows-server 2746:2746
