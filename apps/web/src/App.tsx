import { useEffect, useMemo } from "react";
import "./App.css";
import { socket } from "./modules/socket";

import {
  useLiftSocket,
  useLifts,
  useAddLift,
  useFetchLifts,
  useCallLift,
} from "./modules/liftsv2/store";

type LiftProps = {
  lowestFloor: number;
  highestFloor: number;
  currentFloor: number;
  doorsOpen: boolean;
  onCall: (floor: number) => void;
};

const classNames = (liftAtFloor: boolean, doorsOpen: boolean) => {
  const base = "my-2 p-0";
  if (doorsOpen && liftAtFloor) {
    return base + " bg-green";
  }

  return liftAtFloor ? base + " bg-blue" : base;
};

function Lift(props: LiftProps) {
  const arr = useMemo(
    () =>
      Array.from(
        { length: props.highestFloor - props.lowestFloor + 1 },
        (_, i) => props.highestFloor - i,
      ),
    [props.lowestFloor, props.highestFloor],
  );
  return (
    <div className="flex col mx-2">
      {arr.map(floor => {
        const liftAtFloor = props.currentFloor === floor;
        return (
          <button
            disabled={liftAtFloor}
            key={floor}
            onClick={() => props.onCall(floor)}
            className={classNames(liftAtFloor, props.doorsOpen)}
          >
            {floor}
          </button>
        );
      })}
    </div>
  );
}

const ws = socket("ws://localhost:8080/socket");

function App() {
  useLiftSocket(ws);
  const fetch = useFetchLifts();
  useEffect(() => {
    fetch();
  }, []);

  const lifts = useLifts();
  const callLift = useCallLift();
  const addLift = useAddLift();
  const onAddLift = () => {
    addLift({ floor: 0 });
  };
  return (
    <main>
      <div className="px-4">
        <h1>miffed</h1>
      </div>
      <button onClick={onAddLift}>add lift</button>
      {lifts.map(lift => {
        return (
          <Lift
            key={lift.id}
            lowestFloor={0}
            highestFloor={10}
            doorsOpen={false}
            currentFloor={lift.floor}
            onCall={floor => callLift(lift.id, floor)}
          />
        );
      })}

      <div className="flex align-end"></div>
    </main>
  );
}

export default App;
