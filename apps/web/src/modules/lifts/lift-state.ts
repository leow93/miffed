export type LiftState = {
  currentFloor: number;
  lowestFloor: number;
  highestFloor: number;
  doorsOpen: boolean;
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

type DoorsOpened = {
  type: "lift_doors_opened";
};
const doorsOpened = (): DoorsOpened => ({
  type: "lift_doors_opened",
});
type DoorsClosed = {
  type: "lift_doors_closed";
};
const doorsClosed = (): DoorsClosed => ({
  type: "lift_doors_closed",
});

export const Actions = {
  liftTransited,
  initialise,
  doorsOpened,
  doorsClosed,
};

export type Message = LiftTransited | Initialise | DoorsOpened | DoorsClosed;

export const reducer = (state: State, message: Message): State => {
  if (state.type === "initial") {
    switch (message.type) {
      case "initialise_lift":
        return {
          type: "created",
          currentFloor: message.data.floor,
          lowestFloor: message.data.lowestFloor,
          highestFloor: message.data.highestFloor,
          doorsOpen: false,
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
        doorsOpen: state.doorsOpen,
      };

    case "lift_doors_opened":
      return {
        type: "created",
        currentFloor: state.currentFloor,
        lowestFloor: state.lowestFloor,
        highestFloor: state.highestFloor,
        doorsOpen: true,
      };
    case "lift_doors_closed":
      return {
        type: "created",
        currentFloor: state.currentFloor,
        lowestFloor: state.lowestFloor,
        highestFloor: state.highestFloor,
        doorsOpen: false,
      };
    default:
      return state;
  }
};
