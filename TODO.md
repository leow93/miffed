This is roughly how I'll break the work down:

1. Build a functioning lift model _for a single lift_
2. Build a client/server application that allows a user to interact with the lift
3. Extend our application to support multiple lifts, backend and frontend
4. Make absolutely no profit

# Step 1: Build a functioning lift model 

A lift is a simple thing. It sits there doing nothing until someone presses a button to call it. 
After being called, it needs to move to the correct floor and open its doors to allow passengers to enter. 
Of course, multiple people can call the lift from different floors, and the lift needs to decide which floor to go to next.
This opens up interesting choices about how to optimise the lift's movement:
1. One option is to just move to the floors in order that those floors were called. This is simple to do, but it might not be the most efficient way to move people around.
2. Alternatively, we should probably be a bit smarter about how to move the lift. For example, if the lift is on the ground floor and someone calls it from the 10th floor, it should probably go straight to the 10th floor.
    However, if the lift starts on the ground floor and someone calls the lift from the 10th floor and immediately after someone calls it from the 5th floor, the lift would have less work to do by going to the 5th floor first before arriving at the 10th floor. 

We'll start with option 1, and then very likely move on to option 2.

Motion of a lift is governed both by physics and regulations. We aren't necessarily concerning ourselves with the internal mechanics of a lift, but we do need to consider the following:
- A lift can move at a certain speed. We'll assume that it moves at a constant speed, and that it takes a certain amount of time to move between floors, determined by the speed of the lift and the distance between floors.
- A lift's doors must not open between floors. This would pose all kinds of safety risks, so we'll assume that the lift can only open its doors when it is at a floor.
- A lift may refuse to operate if it is overloaded. We'll assume that the lift has a maximum weight capacity, and that it will refuse to move if the weight of the passengers exceeds this capacity.

Though certainly not exhaustive, these constraints should be enough to get us started.


# Step 2: Build a client/server application
TODO

# Step 3: Extend our application to support multiple lifts
TODO

# Step 4: Make absolutely no profit
Already done