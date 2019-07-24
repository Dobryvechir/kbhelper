package main

import (
	"github.com/Dobryvechir/dvserver/src/dvparser"
	"strings"
)

var template string = `
{
    "kind": "Template",
    "apiVersion": "v1",
    "metadata": {
        "name": "{{MICROSERVICE}}",
        "annotations": {
            "description": "Template for CloudPilot Common Backend Service",
            "tags": "backend",
            "iconClass": "icon-php"
        }
    },
    "parameters": [
        {
            "name": "SERVICE",
            "value": "{{MICROSERVICE}}",
            "description": "Service Name. For example: contact-management-backend",
            "required": false
        },
        {
            "name": "BRANCH",
            "description": "Which git application development branch should be used to deploy",
            "value": "master",
            "required": false
        },
        {
            "name": "TAG",
            "description": "Which docker image tag should be used to deploy",
            "value": "latest",
            "required": false
        },
        {
            "name": "OPENSHIFT_SERVICE_NAME",
            "value": "{{MICROSERVICE}}",
            "description": "Service Name. For example: contact-management-backend",
            "required": false
        },
        {
            "name": "PUBLIC_GATEWAY_URL",
            "description": "Frontend Gateway endpoint url",
            "value": "http://public-gateway-{{OPENSHIFT_PROJECT}}",
            "required": true
        },
        {
            "name": "PRIVATE_GATEWAY_URL",
            "description": "Frontend Gateway endpoint url",
            "value": "http://private-gateway-{{OPENSHIFT_PROJECT}}",
            "required": true
        },
        {
            "name": "PUBLIC_IDENTITY_PROVIDER_URL",
            "description": "Identity Provider endpoint url",
            "value": "http://public-gateway-{{OPENSHIFT_PROJECT}}/api/v1/identity-provider",
            "required": true
        },
        {
            "name": "CERTIFICATE_BUNDLE_MD5SUM",
            "value": "d41d8cd98f00b204e9800998ecf8427e",
            "description": "SSL secret name",
            "required": false
        },
        {
            "name": "SSL_SECRET",
            "value": "defaultsslcertificate",
            "description": "SSL secret name",
            "required": false
        },
        {
            "name": "GLOWROOT_CLUSTER",
            "value": "glowroot.openshift.sdntest.netcracker.com:8181",
            "description": "Host:Port for central glow root",
            "required": false
        }
    ],
    "objects": [
        {
            "kind": "DeploymentConfig",
            "apiVersion": "v1",
            "metadata": {
                "name": "${SERVICE}",
                "labels": {
                    "name": "${SERVICE}"
                }
            },
            "spec": {
                "replicas": 1,
                "strategy": {
                    "type": "Rolling",
                    "rollingParams": {
                        "updatePeriodSeconds": 1,
                        "intervalSeconds": 1,
                        "timeoutSeconds": 600,
                        "maxUnavailable": "25%",
                        "maxSurge": "25%"
                    }
                },
                "template": {
                    "metadata": {
                        "labels": {
                            "name": "${SERVICE}"
                        }
                    },
                    "spec": {
                        "volumes": [
                            {
                                "name": "${SSL_SECRET}",
                                "secret": {
                                    "secretName": "${SSL_SECRET}"
                                }
                            }
                        ],
                        "containers": [
                            {
                                "name": "${SERVICE}",
                                "image": "{{TEMPLATE-IMAGE}}",
                                "volumeMounts": [
                                    {
                                        "name": "${SSL_SECRET}",
                                        "mountPath": "/tmp/cert/${SSL_SECRET}"
                                    }
                                ],
                                "ports": [
                                    {
                                        "containerPort": 8080,
                                        "protocol": "TCP"
                                    }
                                ],
                                "env": [
                                    {
                                        "name": "CERTIFICATE_BUNDLE_${SSL_SECRET}_MD5SUM",
                                        "value": "${CERTIFICATE_BUNDLE_MD5SUM}"
                                    },
                                    {
                                        "name": "PUBLIC_GATEWAY_URL",
                                        "value": "${PUBLIC_GATEWAY_URL}"
                                    },
                                    {
                                        "name": "PRIVATE_GATEWAY_URL",
                                        "value": "${PRIVATE_GATEWAY_URL}"
                                    },
                                    {
                                        "name": "PUBLIC_IDENTITY_PROVIDER_URL",
                                        "value": "${PUBLIC_IDENTITY_PROVIDER_URL}"
                                    },
                                    {
                                        "name": "GLOWROOT_CLUSTER",
                                        "value": "${GLOWROOT_CLUSTER}"
                                    }
                                ],
                                "resources": {
                                    "requests": {
                                        "cpu": "100m",
                                        "memory": "32Mi"
                                    },
                                    "limits": {
                                        "memory": "32Mi",
                                        "cpu": "4"
                                    }
                                }
                            }
                        ]
                    }
                },
                "triggers": [
                    {
                        "type": "ConfigChange"
                    }
                ]
            }
        },
        {
            "kind": "Service",
            "apiVersion": "v1",
            "metadata": {
                "name": "${OPENSHIFT_SERVICE_NAME}"
            },
            "spec": {
                "ports": [
                    {
                        "name": "web",
                        "port": 8080,
                        "targetPort": 8080
                    }
                ],
                "selector": {
                    "name": "${SERVICE}"
                }
            }
        },
        {
            "apiVersion": "v1",
            "kind": "Route",
            "metadata": {
                "name": "{{MICROSERVICE}}"
            },
            "spec": {
                "to": {
                    "kind": "Service",
                    "name": "${OPENSHIFT_SERVICE_NAME}"
                }
            }
        },
        {
            "kind": "ConfigMap",
            "apiVersion": "v1",
            "metadata": {
                "name": "${SERVICE}.monitoring-config"
            },
            "data": {
                "url.health": "http://%(ip)s:8080/health"
            }
        }
    ]
}
`


func composeOpenShiftJsonTemplate(microserviceName string, templateImage string) []byte {
	r := strings.TrimSpace(template)
	host:=dvparser.GlobalProperties["OPENSHIFT_NAMESPACE"]+"."+dvparser.GlobalProperties["OPENSHIFT_SERVER"]+"."+dvparser.GlobalProperties["OPENSHIFT_DOMAIN"]
	replaceMap:=map[string]string {
		"{{MICROSERVICE}}" : microserviceName,
		"{{TEMPLATE-IMAGE}}": templateImage,
		"{{OPENSHIFT_PROJECT}}" : host,

	}
	for k, v:= range replaceMap {
		r = strings.ReplaceAll(r, k, v)
	}
	return []byte(r)
}
