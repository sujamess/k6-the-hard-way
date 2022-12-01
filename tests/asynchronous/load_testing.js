import { check, group, sleep } from 'k6';
import http from 'k6/http';

export const options = {
  stages: [
    { duration: '30s', target: 60 }, // simulate ramp-up of traffic from 1 to 60 users over 30 seconds
    { duration: '30s', target: 60 }, // stay at 60 users for 30 seonds
    { duration: '30s', target: 100 }, // ramp-up to 100 users over 30 seconds (peak hour starts)
    { duration: '30s', target: 100 }, // stay at 100 users for short amount of time (peak hour)
    { duration: '30s', target: 60 }, // ramp-down to 60 users over 30 seconds (peak hour ends)
    { duration: '30s', target: 60 }, // continue at 60 for additional 30 seconds
    { duration: '1m', target: 0 }, // ramp-down to 0 users
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
