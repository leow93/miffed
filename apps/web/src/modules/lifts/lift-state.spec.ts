import { expect, it } from "vitest";
import { LiftState, reducer, Actions } from "./lift-state.ts";

it("initializes the state", () => {
  const state = reducer({ type: "initial" }, Actions.initialise(4, 1, 10));
  expect(state).toEqual({ type: "created", currentFloor: 4, lowestFloor: 1, highestFloor: 10 });
});

it("initialising does nothing if already created", () => {
  const state: LiftState = { type: "created", lowestFloor: 1, highestFloor: 10, currentFloor: 4 };
  const message = Actions.initialise(5, 1, 10);
  const nextState = reducer(state, message);
  expect(nextState).toEqual(state);
});

it("transits do nothing if the state is not created", () => {
  const state: LiftState = { type: "initial" };
  const message = Actions.liftTransited(0, 1);
  const nextState = reducer(state, message);
  expect(nextState).toEqual(state);
});

it("transits to the given floor", () => {
  const state: LiftState = { type: "created", lowestFloor: 1, highestFloor: 10, currentFloor: 1 };
  const message = Actions.liftTransited(1, 2);
  const nextState = reducer(state, message);
  expect(nextState).toEqual({ type: "created", lowestFloor: 1, highestFloor: 10, currentFloor: 2 });
});
