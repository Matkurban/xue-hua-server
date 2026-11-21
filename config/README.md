# 						Server Configuration File Descriptions and Common Configuration Modifications

## External Component Configurations

| Configuration File | Description                                                 |
| ------------------ |-------------------------------------------------------------|
| **kafka.yml**      | Configuration for Kafka username, password, address, etc.   |
| **redis.yml**      | Configuration for Redis password, address, etc.             |
| **rustfs.yml**     | Configuration for RustFS username, password, address, etc.  |
| **mongodb.yml**    | Configuration for MongoDB username, password, address, etc. |
| **discovery.yml**  | Service discovery and etcd credentials and address.         |

## server Related Configurations
| Configuration File              | Description                                    |
| ------------------------------- | ---------------------------------------------- |
| **log.yml**                     | Configuration for logging levels and storage directory                   |
| **notification.yml**            | Event notification settings (e.g., add friend, create group)           |
| **share.yml**                   | Common settings for all services (e.g., secrets)            |
| **webhooks.yml**                | Webhook URLs and related settings                           |
| **local-cache.yml**             | Local cache settings (generally do not modify)                 |
| **server-rpc-third.yml**        | server-rpc-third listen IP, port, and object storage settings  |
| **server-rpc-user.yml**         | server-rpc-user listen IP and port settings              |
| **server-api.yml**              | server-api listen IP, port, and other settings               |
| **server-crontask.yml**         | server-crontask scheduled task settings                   |
| **server-msggateway.yml**       | server-msggateway listen IP, port, and other settings           |
| **server-msgtransfer.yml**      | Settings for server-msgtransfer service                   |
| **server-push.yml**             | server-push listen IP, port, and offline push settings        |
| **server-rpc-auth.yml**         | server-rpc-auth listen IP, port, token validity settings |
| **server-rpc-conversation.yml** | server-rpc-conversation listen IP and port settings     |
| **server-rpc-friend.yml**       | server-rpc-friend listen IP and port settings           |
| **server-rpc-group.yml**        | server-rpc-group listen IP and port settings           |
| **server-rpc-msg.yml**          | server-rpc-msg listen IP and port settings         |


## Monitoring and Alerting Related Configurations
| Configuration File             | Description     |
| ------------------------------ | --------------- |
| **prometheus.yml**             | Prometheus configuration |
| **instance-down-rules.yml**    | Alert rules       |
| **alertmanager.yml**           | Alertmanager configuration   |
| **email.tmpl**                 | Email alert template   |
| **grefana-template/Demo.json** | Default Grafana dashboard |

## Common Configuration Modifications
| Configuration Item                                              | Configuration File                |
| -------------------------------------------------------- | ----------------------- |
| Configure RustFS as object storage (focus on the externalAddress field) | `rustfs.yml`            |
| Adjust log level and number of log files                              | `log.yml`               |
| Enable or disable friend verification when sending messages                           | `server-rpc-msg.yml`    |
| server secret                                         | `share.yml`             |
| Configure OSS, COS, AWS, or Kodo as object storage               | `server-rpc-third.yml`  |
| Multi-end mutual kick strategy and max concurrent connections per gateway                 | `server-msggateway.yml` |
| Offline message push configuration                                            | `server-push.yml`       |
| Configure webhooks for callback notifications (e.g., before/after message send)         | `webhooks.yml`          |
| Whether new group members can view historical messages                          | `server-rpc-group.yml`  |
| Token expiration time settings                                      | `server-rpc-auth.yml`     |
| Scheduled task settings (e.g., how long to retain messages)                      | `server-crontask.yml`   |

## Starting Multiple Instances of a Service and Maximum File Descriptors


To start multiple instances of an Server service, simply add the corresponding port numbers and modify the `start-config.yml` file in the project’s root directory, 
then restart the service. For example, to start 2 instances of `server-rpc-user`:

```yaml
rpc:
  registerIP: ''
  listenIP: 0.0.0.0
  ports: [ 10110, 10111 ]

prometheus:
  enable: true
  ports: [ 20100, 20101 ]
```

Modify`start-config.yml`:

```yaml
serviceBinaries:
  server-rpc-user: 2
```

To set the maximum number of open file descriptors (typically one per online user):

```
maxFileDescriptors: 10000
```
