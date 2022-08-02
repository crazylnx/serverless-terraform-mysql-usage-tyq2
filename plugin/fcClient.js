const { InvokeFunctionRequest } = require('@alicloud/fc-open20210406');
const Client = require('@alicloud/fc-open20210406').default;

class FunctionHelper {
  constructor(config) {
    this.client = new Client(config);
  }

  async invoke(service, functionName, payload) {
    const request = new InvokeFunctionRequest({ body: JSON.stringify(payload) });
    try {
      const response = await this.client.invokeFunction(
        service,
        functionName,
        request,
      );
      return response.body;
    } catch (e) {
      throw new Error(e);
    }
  }
}

module.exports = { FunctionHelper };
