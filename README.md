Namespace
======

#### GET Namespace list
*URL:* **/namespaces**  
*Method:* **GET**  
*Query:*
```
  owner - filter by owner id
```


#### GET Namespace
*URL:* **/namespaces/:namespace/**  
*Method:* **GET**



Deployment
======

#### GET Deploy list
*URL:* **/namespaces/:namespace/deployments**  
*Method:* **GET**
*Query:*
```
  owner - filter by owner id
```

#### GET Deploy
*URL:* **/namespaces/:namespace/deployments/:deployment"**  
*Method:* **GET**



Pods
=====

#### GET Pod list
*URL:* **/namespaces/:namespace/pods**  
*Method:* **GET**
*Query:*
```
  owner - filter by owner id
```

#### GET Pod
*URL:* **/namespaces/:namespace/pods/:pod**  
*Method:* **GET**


Logs
----

#### GET Pod list
*URL:* **/namespaces/:namespace/pods/:pod/logs**  
*Method:* **GET**
*Query:*
```
  follow - bool (default: false)
  tail - int (min: 1, max: 1000)
  container - container name
  previous - bool (default: false)
```
*Extra-Headers*
```
  Sec-Websocket-Version: 13
  Connection: upgrade
  Upgrade: websocket
  Sec-Websocket-Key: 0
```
