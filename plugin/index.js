const core = require('@serverless-devs/core');
const {validate} = require('./validate');
const {FunctionHelper} = require('./fcClient');

const {lodash, Logger} = core;
const logger = new Logger('fc-resource-creator');

/**
 * Plugin 插件入口
 * @param inputs 组件的入口参数
 * @param args 插件的自定义参数 {url: 请求的url, method: 请求方式(默认head), interval: 请求的频率（默认2m）}
 * @return inputs
 */

module.exports = async function index(inputs, args = {}) {
  logger.debug(`inputs params: ${JSON.stringify(inputs)}`);
  logger.debug(`args params: ${JSON.stringify(args)}`);
  const params = {inputs, args};


  validate(params);
  const serviceName = args.service_name;
  const functionName = args.function_name;
  const payload = {
    // invoke_type：0为创建资源，1为删除资源，此处只需要创建。
    invoke_type: 0,
    variables: {
      databases: [{name: args.databaseName, character_set: args.databaseCharacterSet, description: "test"}],
      region: inputs.props.region,
      instance_name: args.instance_name,
      account_name: args.account_name,
      password: args.password,
      allocate_public_connection: true,
      security_ips: ["1.0.0.0/0"],
      privilege: "ReadWrite",
      instance_type: args.instance_type
    },
    oss_config: {
      oss_bucket: args.oss_bucket,
      oss_prefix: args.oss_prefix,
      oss_region: inputs.props.region
    }
  }

  const config = {
    endpoint: `${inputs.credentials.AccountID}.${args.function_region}.fc.aliyuncs.com`,
    accessKeyId: inputs.credentials && inputs.credentials.AccessKeyID,
    accessKeySecret: inputs.credentials && inputs.credentials.AccessKeySecret,
    readTimeout: 1200000,
  };

  const fcClient = new FunctionHelper(config);
  let body;
  await logger.task('Finish', [
    {
      title: 'Invoke resource creator function',
      id: 'invoking function',
      task: async () => {
        body = await fcClient.invoke(
            serviceName,
            functionName,
            payload
        );
      },
    },
  ]);

  const Output = JSON.parse(body.toString());
  if (lodash.get(Output, 'status') != 'SUCCESS') {
    logger.error(`Create resource error, operations: ${Output}`);
    throw new Error(`Create resource error, operations: ${Output}`);
  }
  const result = lodash.get(Output, 'result');
  const terraformOut = JSON.parse(result);
  const resourceConfig = lodash.get(terraformOut, 'outputs');




  const vpc = lodash.get(resourceConfig, 'VPC_ID.value')
  const vSwitch = lodash.get(resourceConfig, 'VSWITCH_ID.value')
  const securityGroup = lodash.get(resourceConfig, 'SECURITY_GROUP_ID.value')

  inputs = lodash.merge(inputs, {
    props: {
      service: {
        vpcConfig: {
          vpcId: vpc,
          securityGroupId: securityGroup,
          vswitchIds: [vSwitch]
        }
      },
      function: {
        environmentVariables: {
          OUTPUT: JSON.stringify(resourceConfig),
        },
      },
    },
  });

  return inputs;
};
