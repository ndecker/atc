package main

var HELP []string = []string{
	`
      Commands:
        <aircraft>A0     aproach airport
        <aircraft>A<1-5> assign altitude
        <aircraft>M      maintain current altitude
        <aircraft>L<0-4> turn left
        <aircraft>R<0-4> turn right
        <aircraft>P      proceed on current heading
        <aircraft>H      hold at navaid
        <aircraft>K      keep current position
        <aircraft><airport>
                         turn towards airport at navaid

        <aircraft>S      status of aircraft

        Esc              quit game
        ,                advance time
        ?                show help
        Tab              show planes`,
	`
      Airplanes:
        Jet (Mark: M)
          A jet is a very common plane that usually
          flys over or starts/lands at an airport.
          It has enough fuel for 15 minutes and needs
          to exit at waypoint markers at level 5

        Propeller aircraft (Mark: P)
          A propeller aircraft is slower than a jet.
          It has time for 21 minutes of flight and flys
          similar routes as the jets.

        Helicopter (Mark: H)
          A helicopter is about as slow as a propeller
          aircraft but it can turn on the spot and
          stay still in the air. It is not usually
          used for long distance travel.

        Blackbird (Mark: B)
          A very rare, fast and high flying plane.
          Usually you should not mess with it.`,
}
