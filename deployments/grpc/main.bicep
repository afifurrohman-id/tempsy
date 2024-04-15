param containerAppName string 
param location string = resourceGroup().location
param image string
param targetPort int
param externalIngress bool = false
param envName string

@maxValue(10)
param maxReplicas int = 2

@allowed([
'auto'
'http'
'http2'
'tcp'
])
param transportProtocol string  = 'auto'

@allowed([
'0.25'
'0.5'
'0.75'
'1.0'
'1.25'
'1.5'
'1.75'
'2.0'
])
param cpuCores string = '0.5'

@allowed([
'0.5Gi'
'1.0Gi' 
'1.5Gi'
'2.0Gi'
'2.5Gi'
'3.0Gi'
'3.5Gi'
'4.0Gi'
])
param memory string =  '1.0Gi'

resource env 'Microsoft.App/managedEnvironments@2023-11-02-preview' = {
  name: envName
  location: location
  properties: {
    workloadProfiles:[
      {
        name: envName
        workloadProfileType:  'consumption' 
      }
    ] 
  }
}


resource containerApp 'Microsoft.App/containerApps@2023-11-02-preview' = {
  name: containerAppName
  location: location
  properties: {
    environmentId: envName == '' ? '' : env.id
    configuration: {
      ingress: {
        external: externalIngress
        targetPort: targetPort
        transport: transportProtocol
        // customDomains: [{
        //   name: ''
        //   bindingType: 'SniEnabled'
        //   certificateId: ''
        // }]
      }
    }
    template: {
      containers: [
        {
          image: image
          name: containerAppName
            env: [
              {
                name:'PORT'
                value: string(targetPort)
              }
            ]
          resources: {
            cpu: json(cpuCores)
            memory: memory
          }
        }
      ]
      scale: {
        maxReplicas: maxReplicas
        minReplicas: 0
      }
    }
  }
}



output containerAppsFQDN string = containerApp.properties.configuration.ingress.fqdn
