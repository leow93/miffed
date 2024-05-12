import { useMemo } from "react";
import "./App.css";
import { socket } from "./modules/socket";
import { useLiftState, useSendMessage } from "./modules/lifts/wiring";

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
  const state = useLiftState(ws);
  const sendMessage = useSendMessage(ws);

  const onCall = (liftId: number, floor: number) => {
    sendMessage({
      liftId,
      type: "call_lift",
      floor,
    });
  };

  return (
    <main>
      <div className="px-4">
        <h1>miffed</h1>
      </div>

      <div className="flex align-end">
        {Object.entries(state).map(([liftId, liftState]) => {
          if (liftState.type === "created") {
            return (
              <Lift
                key={liftId}
                lowestFloor={liftState.lowestFloor}
                highestFloor={liftState.highestFloor}
                currentFloor={liftState.currentFloor}
                doorsOpen={liftState.doorsOpen}
                onCall={floor => onCall(Number(liftId), floor)}
              />
            );
          }

          return null;
        })}
      </div>
    </main>
  );
}

export default App;
