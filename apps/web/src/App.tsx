import { useMemo } from 'react'
import './App.css'

type LiftProps = {
    floors: number
}

function Lift(props: LiftProps) {
    const arr = useMemo(() => Array.from({ length: props.floors }, (_, i) => props.floors - i), [props.floors])
    return (
        <div className="flex col">
            {arr.map(floor => (
                <button key={floor} className="my-2">{floor}</button>
            ))}
        </div>
    )

}

type Lift = {
    id: number
    floor: number
}

function App() {
    return (
        <main>
            <Lift floors={10}/>
        </main>
    )
}

export default App
