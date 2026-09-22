# Flash Sale Lottery with Redis and RabbitMQ and Postgres

A high-concurrency lottery system built using Go, Redis, and MySQL, with weighted prize selection, real-time inventory management, and safe stock deduction to handle high traffic.

## How to run the app

```sh
# 1. start Postgres, Redis and RabbitMQ
docker compose up -d

# 2. wait until all three report healthy (Postgres seeds its tables on first start)
docker compose ps

# 3. run the app
go run .
```

Then open http://localhost:5678/ and spin the wheel

## How to stop the app

Stop the app with Ctrl-C

```sh
docker compose down     # stop the containers, keep the data
docker compose down -v  # stop them and wipe the volumes, so the next start re-seeds Postgres
```

# The flow

Let say three prizes have 5, 2 and 4 units left, and all of them must be handed out.

1. Turn the remaining stock into odds: `probs = (0.45, 0.18, 0.36)`.
2. Accumulate them: `acc = (0.45, 0.64, 1.0)`, which cuts the segment `[0,1]` into three pieces.
3. Draw a random float in `(0,1]`; whichever piece it lands in is the prize.
4. Decrement that prize's stock by one.

Each lottery attempt needs to go through the four steps above again.

Reading and updating inventory both require database operations. In high-concurrency scenarios, Redis can handle much higher QPS than Postgres. So, we use Redis to manage the real-time inventory.

When the inventory is very low, there can be a problem. For example, if only 1 prize is left, 2 goroutines may read the inventory at the same time and both see 1 available. Both may then try to reduce the inventory. Fortunately, Redis’s `DECR` operation is atomic. If the inventory becomes negative after the decrement, it means the operation failed. The goroutine must run the lottery algorithm again.

## Backend API

| Endpoint  | Method | Parameters      | Description                                                         |
| :-------- | :----- | :-------------- | :------------------------------------------------------------------ |
| `/`       | GET    |                 | Returns the lottery wheel page                                      |
| `/gifts`  | GET    |                 | Returns detailed information about all prizes to populate the wheel |
| `/lucky`  | GET    |                 | Returns the ID of the winning prize                                 |
| `/giveup` | POST   | `uid` and `gid` | Give up the prize without making a payment                          |
| `/pay`    | POST   | `uid` and `gid` | Complete the payment                                                |
| `/result` | GET    |                 | Returns the lottery success page                                    |

## Frontend display

The front end is the [lucky-canvas](https://100px.net/usage/js.html) wheel plugin.
