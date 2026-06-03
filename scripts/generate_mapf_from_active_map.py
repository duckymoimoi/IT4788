#!/usr/bin/env python3
import argparse
import json
import random
import re
import subprocess
import sys
from pathlib import Path
from urllib.parse import urlencode
from urllib.request import urlopen


def get_json(url):
    with urlopen(url, timeout=60) as res:
        payload = json.loads(res.read().decode("utf-8"))
    if payload.get("code") != 1000:
        raise RuntimeError(f"{url} returned code={payload.get('code')} message={payload.get('message')}")
    return payload["data"]


def slugify(value):
    value = re.sub(r"[^a-zA-Z0-9]+", "_", value.strip().lower()).strip("_")
    return value or "active_map"


def parse_grid(raw_grid):
    if isinstance(raw_grid, str):
        return json.loads(raw_grid)
    return raw_grid


def write_map_file(path, rows, cols, grid):
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8", newline="\n") as f:
        f.write("type octile\n")
        f.write(f"height {rows}\n")
        f.write(f"width {cols}\n")
        f.write("map\n")
        for row in grid:
            f.write("".join("@" if int(cell) == 1 else "." for cell in row))
            f.write("\n")


def load_active_map_id(base_url):
    data = get_json(f"{base_url}/api/map/sync_full")
    maps = data.get("maps") or []
    if not maps:
        raise RuntimeError("sync_full did not return any active map")
    return int(maps[0]["map_id"])


def sample_agents(grid, rows, cols, poi_locations, count, seed):
    blocked = set(poi_locations)
    walkable = [
        r * cols + c
        for r in range(rows)
        for c in range(cols)
        if int(grid[r][c]) == 0 and (r * cols + c) not in blocked
    ]
    if len(walkable) < count:
        raise RuntimeError(f"not enough walkable cells for {count} agents, got {len(walkable)}")
    rng = random.Random(seed)
    return rng.sample(walkable, count)


def repeat_tasks(poi_locations, count):
    if not poi_locations:
        raise RuntimeError("map has no POI locations to use as MAPF tasks")
    tasks = []
    while len(tasks) < count:
        tasks.extend(poi_locations)
    return tasks[:count]


def write_int_lines(path, values):
    path.parent.mkdir(parents=True, exist_ok=True)
    with path.open("w", encoding="utf-8", newline="\n") as f:
        f.write(str(len(values)))
        f.write("\n")
        for value in values:
            f.write(str(value))
            f.write("\n")


def main():
    parser = argparse.ArgumentParser(description="Generate LoRR MAPF input from hospital active map and POIs.")
    parser.add_argument("--base-url", default="https://group3.it4788.sukkaito.id.vn")
    parser.add_argument("--map-id", type=int, default=0, help="Defaults to current active map from /api/map/sync_full")
    parser.add_argument("--agents", type=int, default=100)
    parser.add_argument("--tasks", type=int, default=300)
    parser.add_argument("--steps", type=int, default=500)
    parser.add_argument("--seed", type=int, default=4788)
    parser.add_argument("--out-dir", default="mapf/generated")
    parser.add_argument("--run", action="store_true", help="Run mapf/build/lifelong.exe after generating input files")
    parser.add_argument("--solver", default="mapf/build/lifelong.exe")
    args = parser.parse_args()

    base_url = args.base_url.rstrip("/")
    map_id = args.map_id or load_active_map_id(base_url)
    meta = get_json(f"{base_url}/api/map/get_meta?{urlencode({'map_id': map_id})}")
    pois = get_json(f"{base_url}/api/map/get_nodes?{urlencode({'map_id': map_id})}")

    rows = int(meta["rows"])
    cols = int(meta["cols"])
    grid = parse_grid(meta["grid_data"])
    if len(grid) != rows or any(len(row) != cols for row in grid):
        raise RuntimeError(f"grid_data shape mismatch: expected {rows}x{cols}")

    poi_locations = []
    off_grid_pois = []
    blocked_pois = []
    for poi in pois:
        loc = int(poi["grid_location"])
        r, c = divmod(loc, cols)
        if not (0 <= r < rows and 0 <= c < cols):
            off_grid_pois.append(poi["poi_code"])
            continue
        if int(grid[r][c]) == 1:
            blocked_pois.append(poi["poi_code"])
            continue
        poi_locations.append(loc)

    if not poi_locations:
        raise RuntimeError("no POI is on a walkable cell; cannot create tasks from POIs")

    name = slugify(meta["map_name"])
    run_dir = Path(args.out_dir) / f"{name}_{map_id}_{args.agents}agents"
    map_path = run_dir / "maps" / f"{name}.map"
    agent_path = run_dir / "agents" / f"{name}_{args.agents}.agents"
    task_path = run_dir / "tasks" / f"{name}_{args.tasks}.tasks"
    input_path = run_dir / f"{name}_{args.agents}.json"
    output_path = run_dir / "output.json"

    agents = sample_agents(grid, rows, cols, poi_locations, args.agents, args.seed)
    tasks = repeat_tasks(poi_locations, args.tasks)
    walkable_count = sum(1 for row in grid for cell in row if int(cell) == 0)

    write_map_file(map_path, rows, cols, grid)
    write_int_lines(agent_path, agents)
    write_int_lines(task_path, tasks)

    input_data = {
        "mapFile": f"maps/{map_path.name}",
        "agentFile": f"agents/{agent_path.name}",
        "teamSize": args.agents,
        "taskFile": f"tasks/{task_path.name}",
        "numTasksReveal": 1,
        "simulation_steps": args.steps,
        "agentCounter": args.agents,
        "version": "2026 LoRR",
        "agentSize": 1.0,
        "enableTaskFinishSpin": True,
    }
    input_path.write_text(json.dumps(input_data, indent=2), encoding="utf-8")

    summary = {
        "map_id": map_id,
        "map_name": meta["map_name"],
        "rows": rows,
        "cols": cols,
        "walkable_cells": walkable_count,
        "wall_cells": rows * cols - walkable_count,
        "poi_count": len(pois),
        "usable_poi_tasks": len(poi_locations),
        "off_grid_pois": off_grid_pois,
        "blocked_pois": blocked_pois,
        "agents": len(agents),
        "tasks": len(tasks),
        "input": str(input_path),
        "output": str(output_path),
    }
    (run_dir / "summary.json").write_text(json.dumps(summary, indent=2, ensure_ascii=False), encoding="utf-8")
    print(json.dumps(summary, indent=2, ensure_ascii=False))

    if args.run:
        solver = Path(args.solver)
        if not solver.exists():
            raise RuntimeError(f"solver not found: {solver}")
        cmd = [
            str(solver.resolve()),
            "--inputFile",
            str(input_path.name),
            "--output",
            str(output_path.name),
            "--simulationTime",
            str(args.steps),
            "--planTimeLimit",
            "1000",
            "--preprocessTimeLimit",
            "30000",
            "--outputScreen",
            "1",
        ]
        completed = subprocess.run(cmd, cwd=run_dir, text=True)
        if completed.returncode != 0:
            raise RuntimeError(f"solver failed with exit code {completed.returncode}")
        print(f"MAPF output written to {output_path}")


if __name__ == "__main__":
    try:
        main()
    except Exception as exc:
        print(f"ERROR: {exc}", file=sys.stderr)
        sys.exit(1)
