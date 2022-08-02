import * as os from "os";

const pkg = {
  name: "@alicloud/credentials",
  version: "2.2.1",
};

export const DEFAULT_UA =
  `AlibabaCloud (${os.platform()}; ${os.arch()}) ` +
  `Node.js/${process.version} Core/${pkg.version}`;

export const DEFAULT_CLIENT = `Node.js(${process.version}), ${pkg.name}: ${pkg.version}`;
