export type LiftState = {
  currentFloor: number;
  lowestFloor: number;
  highestFloor: number;
};

export type State =
  | {
      type: "initial";
    }
  | ({
      type: "created";
    } & LiftState);

export const initialState: State = {
  type: "initial",
};

type LiftTransited = {
  type: "lift_transited";
  data: {
    from: number;
    to: number;
  };
};
const liftTransited = (from: number, to: number): LiftTransited => ({
  type: "lift_transited",
  data: { from, to },
});

type Initialise = {
  type: "initialise_lift";
  data: {
    floor: number;
    lowestFloor: number;
    highestFloor: number;
  };
};
const initialise = (floor: number, lowestFloor: number, highestFloor: number): Initialise => ({
  type: "initialise_lift",
  data: {
    floor,
    lowestFloor,
    highestFloor,
  },
});

export const Actions = {
  liftTransited,
  initialise,
};

export type Message = LiftTransited | Initialise;

export const reducer = (state: State, message: Message): State => {
  if (state.type === "initial") {
    switch (message.type) {
      case "initialise_lift":
        return {
          type: "created",
          currentFloor: message.data.floor,
          lowestFloor: message.data.lowestFloor,
          highestFloor: message.data.highestFloor,
        };
      default:
        return state;
    }
  }

  switch (message.type) {
    case "lift_transited":
      return {
        type: "created",
        currentFloor: message.data.to,
        lowestFloor: state.lowestFloor,
        highestFloor: state.highestFloor,
      };
    default:
      return state;
  }
};
