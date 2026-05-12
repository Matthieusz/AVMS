import { useState, useEffect } from 'react';
import { Scene } from '@/components/Scene';
import { getPhaseLabel, getTimeRemaining, trafficState, type LightColor } from '@/stores/trafficStore';
import { useV2XStore } from '@/stores/v2xStore';
import { IconInfoCircle, IconMap, IconWifi, IconTrafficLights, IconMessageCircle2 } from '@tabler/icons-react';

function LightIndicator({ color, label, path }: { color: LightColor; label: string; path: 'north' | 'south' | 'east' | 'west' }) {
  const colorClasses: Record<LightColor, string> = {
    red: 'bg-red-500 shadow-red-500/50',
    yellow: 'bg-yellow-400 shadow-yellow-400/50',
    green: 'bg-green-500 shadow-green-500/50',
  };

  const dimClasses: Record<LightColor, string> = {
    red: 'bg-red-900/40',
    yellow: 'bg-yellow-900/40',
    green: 'bg-green-900/40',
  };

  const isActive = trafficState[path] === color;

  return (
    <div className="flex items-center gap-1.5">
      <span className="text-[10px] w-8 text-slate-400 uppercase tracking-wide">{label}</span>
      <div
        className={`h-2.5 w-2.5 rounded-full transition-all duration-300 ${
          isActive ? colorClasses[color] : dimClasses[color]
        } ${isActive ? 'shadow-[0_0_6px]' : ''}`}
      />
    </div>
  );
}

function TrafficStatusPanel() {
  const [, setTick] = useState(0);

  useEffect(() => {
    const interval = setInterval(() => setTick((t) => t + 1), 100);
    return () => clearInterval(interval);
  }, []);

  const phase = getPhaseLabel();
  const remaining = getTimeRemaining();

  return (
    <div className="rounded-2xl border border-white/10 bg-slate-950/70 p-4 shadow-xl backdrop-blur-md">
      <div className="mb-3 flex items-center gap-2">
        <IconTrafficLights className="h-4 w-4 text-amber-400" />
        <h2 className="text-sm font-semibold text-white">Traffic Control</h2>
      </div>

      <div className="mb-3 flex items-center justify-between">
        <span className="text-xs text-slate-400">Phase</span>
        <span className="rounded-lg bg-slate-800/80 px-2 py-0.5 text-xs font-medium text-white">
          {phase}
        </span>
      </div>

      <div className="mb-3 flex items-center justify-between">
        <span className="text-xs text-slate-400">Remaining</span>
        <span className="text-xs font-mono text-slate-300">{remaining.toFixed(1)}s</span>
      </div>

      <div className="space-y-1.5 border-t border-white/5 pt-3">
        <div className="flex items-center justify-between">
          <span className="text-[10px] text-slate-500 uppercase tracking-wide">NS Road</span>
          <div className="flex gap-2">
            <LightIndicator color={trafficState.north} label="N" path="north" />
            <LightIndicator color={trafficState.south} label="S" path="south" />
          </div>
        </div>
        <div className="flex items-center justify-between">
          <span className="text-[10px] text-slate-500 uppercase tracking-wide">EW Road</span>
          <div className="flex gap-2">
            <LightIndicator color={trafficState.east} label="E" path="east" />
            <LightIndicator color={trafficState.west} label="W" path="west" />
          </div>
        </div>
      </div>
    </div>
  );
}

function V2XMessagesPanel() {
  const messages = useV2XStore((state) => state.messages);
  const connected = useV2XStore((state) => state.connected);

  return (
    <div className="rounded-2xl border border-white/10 bg-slate-950/70 p-4 shadow-xl backdrop-blur-md h-64 flex flex-col mt-3">
      <div className="mb-3 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <IconMessageCircle2 className="h-4 w-4 text-emerald-400" />
          <h2 className="text-sm font-semibold text-white">V2X Activity</h2>
        </div>
        <div className="flex items-center gap-1.5">
          <div className={`h-2 w-2 rounded-full ${connected ? 'bg-emerald-500 shadow-[0_0_6px_#10b981]' : 'bg-red-500'}`} />
          <span className="text-[10px] text-slate-400 uppercase tracking-wide">
            {connected ? 'WS Connected' : 'Disconnected'}
          </span>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto space-y-2 pr-1 custom-scrollbar text-xs">
        {messages.length === 0 ? (
          <p className="text-slate-500 text-center mt-6">No messages yet...</p>
        ) : (
          messages.map((m) => (
            <div key={m.id} className="bg-slate-900/50 p-2 rounded border border-white/5">
              <div className="flex justify-between items-center mb-1">
                <span className="text-blue-400 font-mono">[{m.carId}]</span>
                <span className="text-slate-400 text-[10px]">
                  {m.timestamp.toLocaleTimeString()}
                </span>
              </div>
              <div className="flex gap-2 items-start">
                <span className={`text-[10px] px-1.5 py-0.5 rounded font-mono ${m.direction === 'in' ? 'bg-emerald-500/20 text-emerald-300' : 'bg-amber-500/20 text-amber-300'}`}>
                  {m.direction === 'in' ? '↓' : '↑'} {m.type}
                </span>
                <span className="text-slate-300 flex-1 truncate" title={m.payload}>
                  {m.payload}
                </span>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}

function App() {
  const [showInfo, setShowInfo] = useState(true);
  const connectV2X = useV2XStore((state) => state.connect);

  useEffect(() => {
    connectV2X();
  }, [connectV2X]);

  return (
    <div className="relative h-screen w-screen overflow-hidden bg-slate-900">
      <Scene />

      {/* Header overlay */}
      <div className="pointer-events-none absolute top-0 left-0 right-0 flex items-start justify-between p-5">
        <div className="pointer-events-auto rounded-2xl border border-white/10 bg-slate-950/70 px-5 py-4 shadow-xl backdrop-blur-md">
          <div className="flex items-center gap-2.5">
            <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-emerald-500/15">
              <IconMap className="h-5 w-5 text-emerald-400" />
            </div>
            <div>
              <h1 className="text-base font-semibold tracking-tight text-white">
                AVMS Simulation
              </h1>
              <p className="text-xs text-slate-400">
                Isometric V2X junction view
              </p>
            </div>
          </div>
        </div>

        <button
          onClick={() => setShowInfo((s) => !s)}
          className="pointer-events-auto flex h-10 w-10 items-center justify-center rounded-xl border border-white/10 bg-slate-950/70 text-slate-300 shadow-lg backdrop-blur-md transition-colors hover:bg-slate-900/90 hover:text-white"
          aria-label="Toggle info"
        >
          <IconInfoCircle className="h-5 w-5" />
        </button>
      </div>

      {/* Info panels */}
      {showInfo && (
        <div className="pointer-events-none absolute right-5 bottom-5 flex max-w-sm flex-col gap-3 w-80">
          <TrafficStatusPanel />
          <V2XMessagesPanel />

          <div className="pointer-events-auto rounded-2xl border border-white/10 bg-slate-950/70 p-4 shadow-xl backdrop-blur-md">
            <div className="mb-2 flex items-center gap-2">
              <IconWifi className="h-4 w-4 text-sky-400" />
              <h2 className="text-sm font-semibold text-white">RSU Coverage</h2>
            </div>
            <p className="text-xs leading-relaxed text-slate-400">
              4 Roadside Units (RSU) are positioned at the junction corners.
              The blue pulses indicate active V2X broadcast range.
            </p>
            <div className="mt-2 flex flex-wrap gap-2">
              <span className="rounded-lg bg-emerald-500/10 px-2 py-1 text-[10px] font-medium text-emerald-400">
                8 Vehicles
              </span>
              <span className="rounded-lg bg-sky-500/10 px-2 py-1 text-[10px] font-medium text-sky-400">
                4 RSUs
              </span>
              <span className="rounded-lg bg-amber-500/10 px-2 py-1 text-[10px] font-medium text-amber-400">
                4 Traffic Lights
              </span>
            </div>
            <p className="mt-2 text-[10px] text-slate-500">
              Drag to pan • Scroll to zoom
            </p>
          </div>
        </div>
      )}
    </div>
  );
}

export default App;
