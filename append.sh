#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<EOF
Usage:
  $0 -file <path> <EventKind> [options] [positional]

Event kinds:
  MapStart | EventMapStart
  MapEnd   | EventMapEnd
  BagInit  | EventBagInit
  BagMod   | EventBagMod

Options:
  -f, -file, --file <path>    Output log file (required)
  --itemId <id>                Item/ConfigBaseId for bag events (required for BagInit/BagMod)
  --next <ScenePath>           Override NextSceneName for MapStart/MapEnd
  --last <ScenePath>           Override LastSceneName for MapStart/MapEnd
  -h, --help                   Show this help

Notes:
- For BagInit/BagMod, you can also pass a single positional numeric argument
  (e.g., "12345") which will be treated as --itemId if --itemId isn't provided.
- PageId, SlotId, and Num default to 100, 1, and 1 respectively.
- The output format matches UE-style logs so the Go parser recognizes them, e.g.:
  [2025.11.04-19.23.40:547][  5]GameLog: Display: [Game] BagMgr@:Modfy BagItem PageId = 103 SlotId = 60 ConfigBaseId = 10042 Num = 34

Examples:
  $0 -file mylog.log MapStart
  $0 -file mylog.log MapEnd
  $0 -file mylog.log BagInit --itemId 3001
  $0 -file mylog.log BagMod --itemId 10042
  # Positional item id is also accepted for bag events:
  $0 -file mylog.log BagInit 3001
  $0 -file mylog.log BagMod 10042
EOF
}

if [[ $# -lt 2 ]]; then
  usage
  exit 1
fi

outfile=""
raw_kind=""
kind=""

# Defaults for Bag* events (can be extended later if needed)
pageId=100
slotId=1
num=1
itemId=""

# Scene path defaults
DEFAULT_REFUGE="/Game/Art/Maps/01SD/XZ_YuJinZhiXiBiNanSuo200/XZ_YuJinZhiXiBiNanSuo200.XZ_YuJinZhiXiBiNanSuo200"
DEFAULT_NON_REFUGE="/Game/Art/Maps/07YJ/YJ_YongZhouHuiLang200/YJ_YongZhouHuiLang200.YJ_YongZhouHuiLang200"
nextPath=""
lastPath=""

# Collect unknown tokens to support positional item id
positional=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    -f|-file|--file)
      if [[ $# -lt 2 ]]; then
        echo "Error: -file requires a path" >&2
        exit 1
      fi
      outfile="$2"; shift 2 ;;
    MapStart|EventMapStart|MapEnd|EventMapEnd|BagInit|EventBagInit|BagMod|EventBagMod)
      raw_kind="$1"; shift ;;
    --itemId)
      if [[ $# -lt 2 ]]; then
        echo "Error: --itemId requires a numeric id" >&2
        exit 1
      fi
      itemId="$2"; shift 2 ;;
    --next)
      if [[ $# -lt 2 ]]; then
        echo "Error: --next requires a scene path" >&2
        exit 1
      fi
      nextPath="$2"; shift 2 ;;
    --last)
      if [[ $# -lt 2 ]]; then
        echo "Error: --last requires a scene path" >&2
        exit 1
      fi
      lastPath="$2"; shift 2 ;;
    -h|--help)
      usage; exit 0 ;;
    *)
      positional+=("$1"); shift ;;
  esac
done

if [[ -z "$outfile" ]]; then
  echo "Error: output file not specified (-file)" >&2
  usage; exit 1
fi
if [[ -z "$raw_kind" ]]; then
  echo "Error: EventKind not specified" >&2
  usage; exit 1
fi

# Normalize kind
case "$raw_kind" in
  MapStart|EventMapStart) kind="MapStart" ;;
  MapEnd|EventMapEnd)     kind="MapEnd"   ;;
  BagInit|EventBagInit)   kind="BagInit"  ;;
  BagMod|EventBagMod)     kind="BagMod"   ;;
  *) echo "Error: unknown EventKind '$raw_kind'" >&2; usage; exit 1 ;;
esac

# Treat first positional numeric as --itemId for bag events if not provided
if [[ ("$kind" == "BagInit" || "$kind" == "BagMod") && -z "$itemId" && ${#positional[@]} -gt 0 ]]; then
  if [[ "${positional[0]}" =~ ^[0-9]+$ ]]; then
    itemId="${positional[0]}"
  else
    echo "Error: expected numeric item id (use --itemId)" >&2
    exit 1
  fi
fi

# Timestamp prefix (UE-style). Use local time; ms is synthesized for portability.
ts_date="$(date '+%Y.%m.%d-%H.%M.%S')"
ms=$(printf "%03d" $((RANDOM % 1000)))
tid=$(( (RANDOM % 999) + 1 ))
tidp=$(printf "%3d" "${tid}")
prefix="[${ts_date}:${ms}][${tidp}]GameLog: Display: [Game] "

# Ensure output directory exists
mkdir -p "$(dirname -- "$outfile")"

# Build body according to kind
body=""
case "$kind" in
  MapStart)
    # defaults: last = refuge, next = non-refuge
    last="${lastPath:-$DEFAULT_REFUGE}"
    next="${nextPath:-$DEFAULT_NON_REFUGE}"
    body="PageApplyBase@ _UpdateGameEnd: LastSceneName = World'${last}' NextSceneName = World'${next}'"
    ;;
  MapEnd)
    # defaults: last = non-refuge, next = refuge
    last="${lastPath:-$DEFAULT_NON_REFUGE}"
    next="${nextPath:-$DEFAULT_REFUGE}"
    body="PageApplyBase@ _UpdateGameEnd: LastSceneName = World'${last}' NextSceneName = World'${next}'"
    ;;
  BagInit)
    if [[ -z "$itemId" ]]; then
      echo "Error: --itemId is required for BagInit (or provide positional numeric id)" >&2
      exit 1
    fi
    body="BagMgr@:InitBagData PageId = ${pageId} SlotId = ${slotId} ConfigBaseId = ${itemId} Num = ${num}"
    ;;
  BagMod)
    if [[ -z "$itemId" ]]; then
      echo "Error: --itemId is required for BagMod (or provide positional numeric id)" >&2
      exit 1
    fi
    body="BagMgr@:Modfy BagItem PageId = ${pageId} SlotId = ${slotId} ConfigBaseId = ${itemId} Num = ${num}"
    ;;
  *)
    echo "Internal error: unsupported kind '$kind'" >&2
    exit 1
    ;;
esac

# Append the final line
printf "%s%s\n" "$prefix" "$body" >> "$outfile"

echo "Appended: ${kind} -> ${outfile}"
