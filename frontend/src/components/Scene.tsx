import { Canvas } from '@react-three/fiber';
import { OrbitControls, OrthographicCamera, Environment } from '@react-three/drei';
import { Ground } from './Ground';
import { StreetGrid } from './StreetGrid';
import { Car } from './Car';
import { RSU } from './RSU';
import { TrafficLight } from './TrafficLight';
import { TrafficController } from './TrafficController';

export function Scene() {
  return (
    <Canvas shadows style={{ width: '100%', height: '100vh' }}>
      <OrthographicCamera
        makeDefault
        position={[30, 30, 30]}
        zoom={25}
        near={0.1}
        far={1000}
      />
      <OrbitControls
        enablePan={true}
        enableZoom={true}
        enableRotate={false}
        minZoom={10}
        maxZoom={80}
        target={[0, 0, 0]}
      />

      <ambientLight intensity={0.5} />
      <directionalLight
        position={[20, 30, 15]}
        intensity={1.5}
        castShadow
        shadow-mapSize-width={2048}
        shadow-mapSize-height={2048}
        shadow-camera-left={-40}
        shadow-camera-right={40}
        shadow-camera-top={40}
        shadow-camera-bottom={-40}
      />
      <directionalLight position={[-10, 20, -10]} intensity={0.4} color="#bfdbfe" />

      <Environment preset="city" environmentIntensity={0.3} />

      <fog attach="fog" args={['#e0f2fe', 50, 90]} />

      <TrafficController />

      <Ground />
      <StreetGrid />

      {/* Traffic lights at each approach */}
      <TrafficLight position={[4, 0, 5.5]} rotation={[0, 0, 0]} path="north" />
      <TrafficLight position={[-4, 0, -5.5]} rotation={[0, Math.PI, 0]} path="south" />
      <TrafficLight position={[-5.5, 0, 4]} rotation={[0, -Math.PI / 2, 0]} path="east" />
      <TrafficLight position={[5.5, 0, -4]} rotation={[0, Math.PI / 2, 0]} path="west" />

      {/* RSUs at each corner of the junction */}
      <RSU position={[-8, 0, -8]} />
      <RSU position={[8, 0, -8]} />
      <RSU position={[-8, 0, 8]} />
      <RSU position={[8, 0, 8]} />

      {/* Cars on North road — left lane only */}
      <Car vehicleId="obu-n1" path="north" lane="left" speed={3.5} color="#ef4444" startOffset={0} />
      <Car vehicleId="obu-n2" path="north" lane="left" speed={3.2} color="#b91c1c" startOffset={30} />

      {/* Cars on South road — left lane only */}
      <Car vehicleId="obu-s1" path="south" lane="left" speed={3.8} color="#10b981" startOffset={8} />
      <Car vehicleId="obu-s2" path="south" lane="left" speed={3.4} color="#059669" startOffset={38} />

      {/* Cars on East road — left lane only */}
      <Car vehicleId="obu-e1" path="east" lane="left" speed={3.2} color="#8b5cf6" startOffset={15} />
      <Car vehicleId="obu-e2" path="east" lane="left" speed={2.8} color="#7c3aed" startOffset={45} />

      {/* Cars on West road — left lane only */}
      <Car vehicleId="obu-w1" path="west" lane="left" speed={3.6} color="#06b6d4" startOffset={22} />
      <Car vehicleId="obu-w2" path="west" lane="left" speed={3.2} color="#0891b2" startOffset={52} />
    </Canvas>
  );
}
