# go-game-dev

A simple game simulation.

We have 3 components.

- A mock user engine. Which mocks N number of users. N is user input from console. This simulates a yes / no answer with a small delay.

- An API server which listens on /submit and forwards to game engine.

- The game engine evaluates the winner and logs it on console.

# Endpoints

API server :8080/submit
Game engine: 9090/game/response