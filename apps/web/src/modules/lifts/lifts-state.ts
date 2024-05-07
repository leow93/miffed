import { LiftState, Message, reducer as liftReducer, State } from "./lift-state.ts";

type LiftsState = Record<string, State>;

type Action =
  | {
      type: "initialise";
      data: Record<string, LiftState>;
    }
  | (Message & { liftId: number });

export const initialState: LiftsState = {};

export const reducer = (state: LiftsState, action: Action): LiftsState => {
  if (action.type === "initialise") {
    return Object.keys(action.data).reduce((acc, liftId) => {
      const data = action.data[liftId];
      acc[liftId] = {
        type: "created",
        currentFloor: data.currentFloor,
        lowestFloor: data.lowestFloor,
        highestFloor: data.highestFloor,
      };
      return acc;
    }, {} as LiftsState);
  }

  const liftState = state[action.liftId];
  if (!liftState) {
    return state;
  }
  const newLiftState = liftReducer(liftState, action);
  return { ...state, [action.liftId]: newLiftState };
};
