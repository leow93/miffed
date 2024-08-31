import { useCallback, useEffect } from "react";
import { create } from "zustand";
import { useShallow } from "zustand/react/shallow";
import { getLifts, createLift, callLift } from "./api";
import { EntityState, createEntityState, pushEntity, updateEntityBy } from "./entity-state";

type LiftId = string;
type Lift = {
  id: LiftId;
  floor: number;
};

type LiftsState = EntityState<LiftId, Lift>;
type StoreState = {
  lifts: LiftsState;
  fetchLifts: () => Promise<void>;
  addLift: (opts: { floor: number }) => Promise<void>;
  callLift: (id: LiftId, floor: number) => Promise<void>;
  changeFloor: (id: LiftId, floor: number) => void;
};

export const useLiftsStore = create<StoreState>((set, get) => ({
  lifts: createEntityState([]),
  fetchLifts: async () => {
    console.log("fetching lifts!");
    const lifts = await getLifts();
    console.log("lifts fetched!");
    set({ lifts: createEntityState(lifts) });
  },
  addLift: async (opts: { floor: number }) => {
    const lift = await createLift(opts);
    set({
      lifts: pushEntity(get().lifts, lift),
    });
  },
  callLift: async (id: LiftId, floor: number) => {
    await callLift(id, floor);
  },
  changeFloor: (id: LiftId, floor: number) => {
    set({
      lifts: updateEntityBy(get().lifts, id, lift => ({ ...lift, floor })),
    });
  },
}));

const keepMap = <T, R>(xs: T[], f: (x: T) => R | null) => {
  const result: R[] = [];
  for (const x of xs) {
    const y = f(x);
    if (y != undefined) {
      result.push(y);
    }
  }
  return result;
};

export const useLifts = () =>
  useLiftsStore(
    useShallow(state =>
      keepMap(state.lifts.ids, id => {
        const lift = state.lifts.entities[id];
        return lift ?? null;
      }),
    ),
  );
export const useFetchLifts = () => useLiftsStore(x => x.fetchLifts);
export const useAddLift = () => useLiftsStore(x => x.addLift);
export const useCallLift = () => useLiftsStore(x => x.callLift);
const useChangeLiftFloor = () => useLiftsStore(x => x.changeFloor);

type LiftEvent =
  | {
      event_type: "lift_transited";
      data: {
        from: number;
        to: number;
      };
      lift_id: LiftId;
    }
  | {
      event_type: "lift_arrived";
      data: {
        floor: number;
      };

      lift_id: LiftId;
    };

export const useLiftSocket = (socket: WebSocket) => {
  const updateLiftFloor = useChangeLiftFloor();
  const handler = useCallback(
    (msg: LiftEvent) => {
      switch (msg.event_type) {
        case "lift_transited":
          updateLiftFloor(msg.lift_id, msg.data.to);
          return;
        case "lift_arrived":
          updateLiftFloor(msg.lift_id, msg.data.floor);
          return;
      }
    },
    [updateLiftFloor],
  );

  useEffect(() => {
    const onMessage = (ev: MessageEvent) => {
      const msg: LiftEvent = JSON.parse(ev.data);
      handler(msg);
    };
    socket.addEventListener("message", onMessage);

    return () => {
      socket.removeEventListener("message", onMessage);
    };
  }, [handler]);
};
