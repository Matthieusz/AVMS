import { useFrame } from '@react-three/fiber';
import { updateTrafficState } from '@/stores/trafficStore';

export function TrafficController() {
  useFrame((_, delta) => {
    updateTrafficState(delta);
  });
  return null;
}
