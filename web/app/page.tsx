"use client";

import { useEffect, useState } from "react";

type Coordinates = { latitude: number; longitude: number };
type LocationStatus = "loading" | "ready" | "unavailable";
type Forecast = {
  Latitude: number;
  Longitude: number;
  Hourly: {
    Time: string[];
    Temperature: number[];
  };
};

export default function Home() {
  const [coords, setCoords] = useState<Coordinates>({
    latitude: 0,
    longitude: 0,
  });
  const [locationStatus, setLocationStatus] = useState<LocationStatus>(() =>
    typeof navigator !== "undefined" && navigator.geolocation
      ? "loading"
      : "unavailable",
  );
  const [locationMessage, setLocationMessage] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [result, setResult] = useState<Forecast | null>(null);

  useEffect(() => {
    if (!navigator.geolocation) {
      return;
    }

    navigator.geolocation.getCurrentPosition(
      (position) => {
        setCoords({
          latitude: Number(position.coords.latitude.toFixed(6)),
          longitude: Number(position.coords.longitude.toFixed(6)),
        });
        setLocationMessage(null);
        setLocationStatus("ready");
      },
      (geolocationError) => {
        setLocationMessage(geolocationError.message);
        setLocationStatus("unavailable");
      },
      {
        enableHighAccuracy: false,
        timeout: 20000,
        maximumAge: 300000,
      },
    );
  }, []);

  async function handleSubmit(event: React.SubmitEvent<HTMLFormElement>) {
    event.preventDefault();

    const { latitude, longitude } = coords;
    setIsSubmitting(true);
    setError(null);
    setResult(null);
    const apiBaseUrl = process.env.NEXT_PUBLIC_API_BASE_URL;

    if (!apiBaseUrl) {
      setError("API base URL not set");
      setIsSubmitting(false);
      return;
    }

    fetch(`${apiBaseUrl}/weather?lat=${latitude}&lon=${longitude}`)
      .then(async (res) => {
        const data = await res.json();

        if (!res.ok) {
          throw new Error(data.error ?? "Failed to fetch forecast");
        }

        return data as Forecast;
      })
      .then((data) => {
        setResult(data);
        setIsSubmitting(false);
      })
      .catch((err) => {
        setError(err.message);
        setIsSubmitting(false);
      });
  }

  const hourlyForecast = result
    ? result.Hourly.Time.map((time, index) => ({
        time,
        temperature: result.Hourly.Temperature[index],
      }))
    : [];
  const temperatures = hourlyForecast
    .map(({ temperature }) => temperature)
    .filter((temperature) => typeof temperature === "number");
  const minTemperature = temperatures.length ? Math.min(...temperatures) : 0;
  const maxTemperature = temperatures.length ? Math.max(...temperatures) : 0;
  const temperatureRange = Math.max(maxTemperature - minTemperature, 1);
  const chartPoints = hourlyForecast
    .map(({ temperature }, index) => {
      const x =
        hourlyForecast.length === 1
          ? 0
          : (index / (hourlyForecast.length - 1)) * 100;
      const y = 100 - ((temperature - minTemperature) / temperatureRange) * 100;
      return `${x.toFixed(2)},${y.toFixed(2)}`;
    })
    .join(" ");
  const xAxisLabels = hourlyForecast.filter((_, index) => {
    if (hourlyForecast.length <= 6) {
      return true;
    }
    const step = Math.ceil((hourlyForecast.length - 1) / 5);
    return index % step === 0 || index === hourlyForecast.length - 1;
  });
  const yAxisLabels = [
    maxTemperature,
    (maxTemperature + minTemperature) / 2,
    minTemperature,
  ];

  return (
    <main className="min-h-screen px-5 py-8 text-foreground sm:px-8 lg:px-12">
      <section className="mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-6xl flex-col justify-center gap-8">
        <div className="max-w-3xl">
          <p className="mb-4 font-mono text-xs font-semibold uppercase tracking-[0.35em] text-clay">
            coordinate forecast
          </p>
          <h1 className="text-5xl font-semibold leading-[0.95] tracking-[-0.06em] sm:text-7xl">
            Weather by exact latitude and longitude.
          </h1>
          <p className="mt-5 max-w-2xl text-lg leading-8 text-ink-muted">
            Enter coordinates and get the full hourly temperature series from
            the backend forecast response.
          </p>
        </div>

        <div className="grid gap-6 lg:grid-cols-[minmax(0,0.9fr)_minmax(0,1.1fr)]">
          <form
            onSubmit={handleSubmit}
            className="rounded-4xl border border-ring bg-paper/85 p-6 shadow-[0_24px_80px_rgba(38,63,43,0.16)] backdrop-blur sm:p-8"
          >
            <div className="mb-8 flex items-center justify-between gap-4">
              <div>
                <h2 className="text-2xl font-semibold tracking-[-0.04em]">
                  Search
                </h2>
                <p className="mt-1 text-sm text-ink-muted">
                  {locationStatus === "loading" &&
                    "Requesting your browser location..."}
                  {locationStatus === "ready" &&
                    "Using your browser location by default."}
                  {locationStatus === "unavailable" &&
                    `Location unavailable${locationMessage ? `: ${locationMessage}` : ""}. Enter coordinates manually.`}
                </p>
              </div>
              <span className="rounded-full bg-moss-dark px-3 py-1 font-mono text-xs font-semibold text-paper">
                GET /weather
              </span>
            </div>

            <div className="grid gap-4">
              <label className="grid gap-2">
                <span className="font-mono text-xs font-semibold uppercase tracking-[0.2em] text-ink-muted">
                  Latitude
                </span>
                <input
                  className="rounded-2xl border border-ring bg-white/70 px-4 py-3 text-lg outline-none transition focus:border-moss focus:ring-4 focus:ring-ring"
                  type="number"
                  step="any"
                  value={coords.latitude}
                  onChange={(e) =>
                    setCoords({ ...coords, latitude: Number(e.target.value) })
                  }
                />
              </label>

              <label className="grid gap-2">
                <span className="font-mono text-xs font-semibold uppercase tracking-[0.2em] text-ink-muted">
                  Longitude
                </span>
                <input
                  className="rounded-2xl border border-ring bg-white/70 px-4 py-3 text-lg outline-none transition focus:border-moss focus:ring-4 focus:ring-ring"
                  type="number"
                  step="any"
                  value={coords.longitude}
                  onChange={(e) =>
                    setCoords({ ...coords, longitude: Number(e.target.value) })
                  }
                />
              </label>
            </div>

            <button
              className="mt-6 w-full rounded-2xl bg-moss-dark px-5 py-4 font-mono text-sm font-semibold uppercase tracking-[0.22em] text-paper transition hover:-translate-y-0.5 hover:bg-moss disabled:cursor-not-allowed disabled:opacity-60 disabled:hover:translate-y-0"
              type="submit"
              disabled={isSubmitting}
            >
              {isSubmitting ? "Loading forecast" : "Fetch forecast"}
            </button>

            {error && (
              <p className="mt-4 rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm font-medium text-red-700">
                {error}
              </p>
            )}
          </form>

          <section className="min-h-112 rounded-4xl border border-ring bg-[#17211b] p-6 text-paper shadow-[0_24px_80px_rgba(23,33,27,0.2)] sm:p-8">
            {!result && !isSubmitting && (
              <div className="flex h-full flex-col justify-between">
                <div>
                  <p className="font-mono text-xs font-semibold uppercase tracking-[0.35em] text-sky">
                    waiting
                  </p>
                  <h2 className="mt-4 max-w-md text-4xl font-semibold leading-none tracking-tighter">
                    Forecast data will land here.
                  </h2>
                </div>
                <div className="mt-10 grid grid-cols-3 gap-3 opacity-70">
                  {Array.from({ length: 6 }).map((_, index) => (
                    <div
                      className="h-20 rounded-2xl border border-white/10 bg-white/5"
                      key={index}
                    />
                  ))}
                </div>
              </div>
            )}

            {isSubmitting && (
              <div className="flex h-full items-center justify-center">
                <p className="animate-pulse font-mono text-sm uppercase tracking-[0.3em] text-sky">
                  Fetching weather
                </p>
              </div>
            )}

            {result && (
              <div>
                <p className="font-mono text-xs font-semibold uppercase tracking-[0.35em] text-sky">
                  {result.Latitude.toFixed(4)}, {result.Longitude.toFixed(4)}
                </p>
                <h2 className="mt-4 text-4xl font-semibold leading-none tracking-tighter">
                  Next hours
                </h2>

                <div className="mt-8 rounded-3xl border border-white/10 bg-white/[0.07] p-4">
                  <div className="relative">
                    <div className="grid grid-cols-[3.5rem_minmax(0,1fr)] gap-3">
                      <div className="flex h-56 flex-col justify-between py-1 text-right font-mono text-[0.65rem] text-white/55">
                        {yAxisLabels.map((temperature) => (
                          <span key={temperature}>
                            {temperature.toFixed(1)}°C
                          </span>
                        ))}
                      </div>
                      <svg
                        className="h-56 w-full overflow-visible"
                        preserveAspectRatio="none"
                        role="img"
                        viewBox="0 0 100 100"
                      >
                        <title>Hourly temperature chart</title>
                        {[0, 25, 50, 75, 100].map((y) => (
                          <line
                            className="stroke-white/10"
                            key={y}
                            vectorEffect="non-scaling-stroke"
                            x1="0"
                            x2="100"
                            y1={y}
                            y2={y}
                          />
                        ))}
                        <polyline
                          className="fill-none stroke-sky"
                          points={chartPoints}
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth="3"
                          vectorEffect="non-scaling-stroke"
                        />
                      </svg>
                    </div>
                    <div className="mt-4 ml-17 flex justify-between gap-3 font-mono text-[0.65rem] text-white/55">
                      {xAxisLabels.map(({ time }) => (
                        <span
                          className="min-w-0 truncate"
                          key={time}
                          title={time}
                        >
                          {time.slice(5, 16).replace("T", " ")}
                        </span>
                      ))}
                    </div>
                  </div>
                </div>

                <div className="mt-5 grid grid-cols-3 gap-3">
                  <div className="rounded-2xl border border-white/10 bg-white/[0.07] p-4">
                    <p className="font-mono text-xs text-white/55">Points</p>
                    <p className="mt-2 text-2xl font-semibold tracking-tighter">
                      {hourlyForecast.length}
                    </p>
                  </div>
                  <div className="rounded-2xl border border-white/10 bg-white/[0.07] p-4">
                    <p className="font-mono text-xs text-white/55">Low</p>
                    <p className="mt-2 text-2xl font-semibold tracking-tighter">
                      {minTemperature.toFixed(1)}°C
                    </p>
                  </div>
                  <div className="rounded-2xl border border-white/10 bg-white/[0.07] p-4">
                    <p className="font-mono text-xs text-white/55">High</p>
                    <p className="mt-2 text-2xl font-semibold tracking-tighter">
                      {maxTemperature.toFixed(1)}°C
                    </p>
                  </div>
                </div>
              </div>
            )}
          </section>
        </div>
      </section>
    </main>
  );
}
