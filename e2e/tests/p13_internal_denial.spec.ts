import { expect, test } from '@playwright/test';

test.describe('P13 internal API denial', () => {
  test('/internal is denied through the public proxy and api-server is not host-exposed', async ({ request }) => {
    const publicResponse = await request.get('http://127.0.0.1:19091/internal/posts/1/events');
    expect(publicResponse.status()).toBe(404);

    const directResponse = await request.get('http://127.0.0.1:8080/healthz', {
      failOnStatusCode: false,
      timeout: 1000,
    }).catch((error: Error) => error);

    expect(directResponse).toBeInstanceOf(Error);
  });
});
