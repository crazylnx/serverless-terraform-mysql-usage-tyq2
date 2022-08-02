const core = require('@serverless-devs/core');

const {lodash} = core;

function validate(params) {
  const {args} = params;
  const {inputs} = params;

  if (lodash.isEmpty(inputs.props)) {
    throw new Error(
        'missing props parameter in inputs, please check yaml',
    );
  }
  if (lodash.isEmpty(args.service_name)) {
    throw new Error(
        'missing service_name parameter in fc-resource-creator plugin.',
    );
  }

  if (lodash.isEmpty(args.function_name)) {
    throw new Error(
        'missing function_name parameter in fc-resource-creator plugin.',
    );
  }
  if (lodash.isEmpty(args.function_region)) {
    throw new Error(
        'missing function_region parameter in fc-resource-creator plugin.',
    );
  }


  if (lodash.isEmpty(args.oss_bucket)) {
    throw new Error(
        'missing variables.oss_bucket parameter in fc-resource-creator plugin.',
    );
  }

  if (lodash.isEmpty(args.oss_prefix)) {
    throw new Error(
        'missing variables.oss_prefix parameter in fc-resource-creator plugin.',
    );
  }

}

module.exports = {validate};
