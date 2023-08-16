# ddclient-azure
Dynamic DNS client for Azure DNS Zone

## Please set env:
1. AZURE_RESOURCE_GROUP
2. AZURE_DOMAIN_NAME
3. AZURE_SUBSCRIPTION_ID
4. AZURE_TENANT_ID
5. AZURE_CLIENT_ID
6. AZURE_CLIENT_SECRET or AZURE_CLIENT_CERTIFICATE_PATH

## Docker image

### Push
docker login --username=hangyuxu \
docker build -t {tag} . \
docker tag {ImageId} hangyuxu/ddclient:{tag} \
docker push hangyuxu/ddclient:{tag}

### Pull
docker pull hangyuxu/ddclient:{tag}