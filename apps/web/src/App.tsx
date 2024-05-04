import {useMemo} from "react";
import "./App.css";
import {socket} from "./modules/socket";
import {useLiftState, useSendMessage} from "./modules/lifts/wiring";

type LiftProps = {
    lowestFloor: number;
    highestFloor: number;
    currentFloor: number;
    onCall: (floor: number) => void;
};

const classNames = (currentFloor: number, floor: number) =>
    currentFloor === floor ? "bg-blue my-2" : "my-2";

function Lift(props: LiftProps) {
    const arr = useMemo(
        () =>
            Array.from(
                {length: props.highestFloor - props.lowestFloor + 1},
                (_, i) => props.highestFloor - i,
            ),
        [props.lowestFloor, props.highestFloor],
    );
    return (
        <div className="flex col">
            {arr.map(floor => (
                <button
                    key={floor}
                    onClick={() => props.onCall(floor)}
                    className={classNames(props.currentFloor, floor)}
                >
                    {floor}
                </button>
            ))}
        </div>
    );
}

function App() {
    const ws = useMemo(() => socket("ws://localhost:8080/socket"), []);

    const state = useLiftState(ws);
    const sendMessage = useSendMessage(ws);

    const onCall = (floor: number) => {
        sendMessage({
            type: "call_lift",
            floor,
        });
    };

    return (
        <main>
            {state.type === "created" && (
                <Lift
                    lowestFloor={state.lowestFloor}
                    highestFloor={state.highestFloor}
                    onCall={onCall}
                    currentFloor={state.currentFloor}
                />
            )}
        </main>
    );
}

export default App;
