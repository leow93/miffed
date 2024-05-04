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
        <div className="flex col mx-2">
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

const ws = socket("ws://localhost:8080/socket");
function App() {

    const state = useLiftState(ws);
    const sendMessage = useSendMessage(ws);

    const onCall = (liftId: string, floor: number) => {
        sendMessage({
            liftId,
            type: "call_lift",
            floor,
        });
    };

    return (
        <main>
            <div className='flex align-end'>
                {Object.entries(state).map(([liftId, liftState]) => {
                    if (liftState.type === "created") {
                        return (
                            <Lift
                                key={liftId}
                                lowestFloor={liftState.lowestFloor}
                                highestFloor={liftState.highestFloor}
                                currentFloor={liftState.currentFloor}
                                onCall={floor => onCall(liftId, floor)}
                            />
                        );
                    }

                    return null
                })}
            </div>

        </main>
    );
}

export default App;
