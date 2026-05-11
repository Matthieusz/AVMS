export interface CarEntry {
  id: string;
  path: string;
  lane: string;
  progress: number;
}

const cars = new Map<string, CarEntry>();

export function registerCar(id: string, path: string, lane: string) {
  cars.set(id, { id, path, lane, progress: 0 });
}

export function updateCar(id: string, progress: number) {
  const car = cars.get(id);
  if (car) car.progress = progress;
}

export function unregisterCar(id: string) {
  cars.delete(id);
}

export function getCarAhead(
  id: string,
  path: string,
  lane: string,
): { progress: number; distance: number } | null {
  const myCar = cars.get(id);
  if (!myCar) return null;

  let closestAhead: CarEntry | null = null;
  let minDist = Infinity;
  const range = 60;
  const myWrapped = ((myCar.progress % range) + range) % range;

  for (const car of cars.values()) {
    if (car.id === id) continue;
    if (car.path !== path || car.lane !== lane) continue;

    const otherWrapped = ((car.progress % range) + range) % range;
    let dist = (otherWrapped - myWrapped + range) % range;

    if (dist > range / 2) continue;

    if (dist < minDist) {
      minDist = dist;
      closestAhead = car;
    }
  }

  if (!closestAhead) return null;
  return { progress: closestAhead.progress, distance: minDist };
}
