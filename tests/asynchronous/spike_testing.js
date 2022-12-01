import { check, group, sleep } from 'k6';
import http from 'k6/http';

export const options = {
  stages: [
    { duration: '30s', target: 100 }, // below normal load
    { duration: '30s', target: 100 },
    { duration: '15s', target: 1400 }, // spike to 1400 users
    { duration: '45s', target: 1400 }, // stay at 1400 for 45 seconds
    { duration: '10s', target: 100 }, // scale down. Recovery stage.
    { duration: '50s', target: 100 },
    { duration: '1m', target: 0 },
  ],
};

const SLEEP_DURATION = 0.1;

function getRandomInt(min, max) {
  min = Math.ceil(min);
  max = Math.floor(max);
  return Math.floor(Math.random() * (max - min) + min);
}

export default function() {
  group('asynchronous solution', () => {
    let cartUUID;
    group('create cart', () => {
      const createCartResponse = http.post('http://localhost:3000/carts');
      cartUUID = createCartResponse.json().cartUUID;
      check(createCartResponse, {
        'status code should be 200': createCartResponse => createCartResponse.status === 201,
        'cart uuid must not be empty': cartUUID => cartUUID !== '',
      });
      sleep(SLEEP_DURATION);
    });

    group('add 5 products to cart', () => {
      for (let i = 0; i < 5; i++) {
        const addProductToCartResponse = http.post(
          `http://localhost:3000/carts/${cartUUID}/product`,
          JSON.stringify({ productID: getRandomInt(1, 110000), quantity: getRandomInt(1, 100) }),
        );
        check(addProductToCartResponse, { 'status code should be 201': addProductToCartResponse => addProductToCartResponse.status === 201 });
        sleep(SLEEP_DURATION);
      }
    });

    group('checkout cart', () => {
      const checkoutResponse = http.post(`http://localhost:3000/carts/${cartUUID}/checkout-with-async`);
      check(checkoutResponse, { 'status code should be 200': checkoutResponse => checkoutResponse.status === 201 });
      sleep(SLEEP_DURATION);
    });
  });
}
