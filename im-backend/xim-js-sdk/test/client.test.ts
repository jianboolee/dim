import { describe, it, expect } from 'vitest';
import { MySDK } from '../src/core/client';

describe('MySDK', () => {
  it('should say hello', () => {
    const sdk = new MySDK({ projectName: 'TestProject' });
    expect(sdk.sayHello('ChatGPT')).toBe('Hello, ChatGPT! From TestProject');
  });
});
