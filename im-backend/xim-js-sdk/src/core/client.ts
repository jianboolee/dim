import type { SDKOptions } from '../types';

export class MySDK {
  private options: SDKOptions;

  constructor(options: SDKOptions) {
    this.options = options;
  }

  sayHello(name: string): string {
    return `Hello, ${name}! From ${this.options.projectName}`;
  }
}
