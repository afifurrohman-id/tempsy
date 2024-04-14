param containerAppsName string 
param location string = resourceGroup().location
param environmentId string
param image string 
param targetPort string
param externalIngress bool
param maxReplicas int

resource containerApps 'Microsoft.App/containerApps@2023-11-02-preview' = {
  name: containerAppsName
  location: location
  properties: {
    environmentId: environmentId
    configuration: {
      ingress: {
        external: externalIngress
        targetPort: int(targetPort) 
      }
    }
    template: {
      containers: [
        {
          image: image
          name: containerAppsName
            env: [
              {
                name:'PORT'
                value: targetPort
              }
            ]

        }
      ]
      scale: {
        maxReplicas: maxReplicas
        minReplicas: 0
      }
    }
  }
}


output containerAppsFQDN string = containerApps.properties.configuration.ingress.fqdn
