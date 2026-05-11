export type LightColor = 'red' | 'yellow' | 'green';

export interface TrafficState {
  north: LightColor;
  south: LightColor;
  east: LightColor;
  west: LightColor;
  phase: string;
  timeInPhase: number;
  phaseDuration: number;
}

export const trafficState: TrafficState = {
  north: 'red',
  south: 'red',
  east: 'red',
  west: 'red',
  phase: 'all_red',
  timeInPhase: 0,
  phaseDuration: 2,
};

const PHASE_SEQUENCE = ['all_red', 'ns_green', 'ns_yellow', 'all_red', 'ew_green', 'ew_yellow'];
const PHASE_DURATIONS: Record<string, number> = {
  all_red: 2,
  ns_green: 10,
  ns_yellow: 3,
  ew_green: 10,
  ew_yellow: 3,
};

let currentPhaseIndex = 0;

export function getLightColorForPath(path: string): LightColor {
  return trafficState[path as keyof Omit<TrafficState, 'phase' | 'timeInPhase' | 'phaseDuration'>] as LightColor;
}

export function updateTrafficState(delta: number) {
  trafficState.timeInPhase += delta;

  const phase = PHASE_SEQUENCE[currentPhaseIndex];
  const duration = PHASE_DURATIONS[phase];
  trafficState.phaseDuration = duration;

  if (trafficState.timeInPhase >= duration) {
    currentPhaseIndex = (currentPhaseIndex + 1) % PHASE_SEQUENCE.length;
    const nextPhase = PHASE_SEQUENCE[currentPhaseIndex];
    trafficState.phase = nextPhase;
    trafficState.timeInPhase = 0;
    trafficState.phaseDuration = PHASE_DURATIONS[nextPhase];

    switch (nextPhase) {
      case 'ns_green':
        trafficState.north = 'green';
        trafficState.south = 'green';
        trafficState.east = 'red';
        trafficState.west = 'red';
        break;
      case 'ns_yellow':
        trafficState.north = 'yellow';
        trafficState.south = 'yellow';
        trafficState.east = 'red';
        trafficState.west = 'red';
        break;
      case 'ew_green':
        trafficState.north = 'red';
        trafficState.south = 'red';
        trafficState.east = 'green';
        trafficState.west = 'green';
        break;
      case 'ew_yellow':
        trafficState.north = 'red';
        trafficState.south = 'red';
        trafficState.east = 'yellow';
        trafficState.west = 'yellow';
        break;
      case 'all_red':
        trafficState.north = 'red';
        trafficState.south = 'red';
        trafficState.east = 'red';
        trafficState.west = 'red';
        break;
    }
  }
}

export function getPhaseLabel(): string {
  switch (trafficState.phase) {
    case 'ns_green':
      return 'NS Green';
    case 'ns_yellow':
      return 'NS Yellow';
    case 'ew_green':
      return 'EW Green';
    case 'ew_yellow':
      return 'EW Yellow';
    case 'all_red':
      return 'All Red';
    default:
      return trafficState.phase;
  }
}

export function getTimeRemaining(): number {
  return Math.max(0, trafficState.phaseDuration - trafficState.timeInPhase);
}
