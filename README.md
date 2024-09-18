# AutoCoffee - a NATS based microservices demo

This is a simple demo of a microservices architecture using NATS as the messaging system. The demo is based on a "coffee shop" scenario where clients can order coffee in an automated way.

# Architecture

The demo consists of the following components:
- `routes` - a service that exposes a web interface for clients to order coffee.
- `controller` - a service that receives the orders from the `routes` service, check the stock (using the `stock` service) and send the order to `coffee-makers` services.
- `stock` - a service that keeps track of the stock of coffee beans in a database.
- `coffee-makers` - a service that receives the orders from the `controller` service and makes the coffee. The service updates the stock in the `stock` service when an order is completed.

# Running the demo

You will need to have a NATS server running. 
