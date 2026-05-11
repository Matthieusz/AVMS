import { useRef } from 'react';
import { useFrame } from '@react-three/fiber';
import { Box, Sphere, Cylinder } from '@react-three/drei';
import * as THREE from 'three';

interface RSUProps {
  position: [number, number, number];
  label?: string;
}

export function RSU({ position }: RSUProps) {
  const pulseRef = useRef<THREE.Mesh>(null);
  const pulseRef2 = useRef<THREE.Mesh>(null);

  useFrame(({ clock }) => {
    const t = clock.getElapsedTime();
    if (pulseRef.current) {
      const scale = 1 + Math.sin(t * 3) * 0.15;
      pulseRef.current.scale.setScalar(scale);
      const mat = pulseRef.current.material as THREE.MeshStandardMaterial;
      mat.opacity = 0.4 + Math.sin(t * 3) * 0.3;
    }
    if (pulseRef2.current) {
      const scale = 1 + Math.sin(t * 3 + 1) * 0.2;
      pulseRef2.current.scale.setScalar(scale);
      const mat = pulseRef2.current.material as THREE.MeshStandardMaterial;
      mat.opacity = 0.2 + Math.sin(t * 3 + 1) * 0.15;
    }
  });

  return (
    <group position={position}>
      {/* Pole */}
      <Cylinder args={[0.08, 0.1, 4]} position={[0, 2, 0]} castShadow>
        <meshStandardMaterial color="#6b7280" metalness={0.6} roughness={0.3} />
      </Cylinder>

      {/* RSU Box */}
      <Box args={[0.8, 0.5, 0.4]} position={[0, 4.1, 0]} castShadow>
        <meshStandardMaterial color="#f97316" metalness={0.3} roughness={0.4} />
      </Box>

      {/* Antenna */}
      <Cylinder args={[0.02, 0.02, 0.6]} position={[0, 4.6, 0]}>
        <meshStandardMaterial color="#374151" metalness={0.7} />
      </Cylinder>
      <Sphere args={[0.06]} position={[0, 4.95, 0]}>
        <meshStandardMaterial color="#ef4444" emissive="#ef4444" emissiveIntensity={2} />
      </Sphere>

      {/* Signal pulses */}
      <Sphere ref={pulseRef} args={[1.2]} position={[0, 4.3, 0]}>
        <meshStandardMaterial color="#3b82f6" transparent opacity={0.4} />
      </Sphere>
      <Sphere ref={pulseRef2} args={[2]} position={[0, 4.3, 0]}>
        <meshStandardMaterial color="#3b82f6" transparent opacity={0.2} />
      </Sphere>

      {/* Base plate */}
      <Cylinder args={[0.4, 0.5, 0.15]} position={[0, 0.075, 0]} receiveShadow>
        <meshStandardMaterial color="#4b5563" />
      </Cylinder>
    </group>
  );
}
