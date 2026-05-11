import { Plane } from '@react-three/drei';

export function Ground() {
  return (
    <Plane args={[80, 80]} rotation={[-Math.PI / 2, 0, 0]} position={[0, -0.01, 0]} receiveShadow>
      <meshStandardMaterial color="#c8d6b9" />
    </Plane>
  );
}
