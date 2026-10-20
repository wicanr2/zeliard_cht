#!/bin/sh
# Docker內使用；輸出根由外層明確掛載，不覆寫既有收據。
set -eu
export GOWORK=/tmp/cave-platform.work
printf 'go 1.24.0\nuse (\n/project\n/project/workplace/music-release-v1/release-build/dosgolem\n)\n' > "$GOWORK"
export ZELIARD_PROJECT=/project
export ZELIARD_ORIG=/project/workplace/orig
export ZELIARD_CAVE_STUDY=/project/workplace/cave-hd-v1
export ZELIARD_FONT=/project/workplace/music-release-v1/release-build/AppDir/font.hex.gz
export ZELIARD_CAVE_ART=/project/workplace/music-release-v1/release-build/AppDir/assets/cave
export ZELIARD_CAVE_PLAYER_ART="$ZELIARD_CAVE_ART"
export ZELIARD_CAVE_WEAPON_ART="$ZELIARD_CAVE_ART"
export ZELIARD_CAVE_OBJECT_ART="$ZELIARD_CAVE_ART"
base=/project/workplace/full-hd-v1
export ZELIARD_CAVE_ICE_ART="$base/cave-ice-complete-v1"
export ZELIARD_CAVE_PELIGRO_ART="$base/cave-peligro-complete-v1"
export ZELIARD_CAVE_MPP3_ART="$base/cave-mpp3-complete-v2"
export ZELIARD_CAVE_MPP5_ART="$base/cave-mpp5-complete-v1"
export ZELIARD_CAVE_MPP6_ART="$base/cave-mpp6-complete-v2"
export ZELIARD_CAVE_MPP7_ART="$base/cave-mpp7-complete-v1"
export ZELIARD_CAVE_MPP8_ART="$base/cave-mpp8-complete-v1"
export ZELIARD_CAVE_MPP9_ART="$base/cave-mpp9-complete-v1"
export ZELIARD_CAVE_MPPA_ART="$base/cave-mppa-complete-v1"
export ZELIARD_CAVE_MPPB_ART="$base/cave-mppb-complete-v1"
export ZELIARD_CAVE_MPP1_ART="$base/cave-mpp1-complete-v2"
export ZELIARD_CAVE_COMMON_ART="$base/cave-common-registration-v2"
export ZELIARD_CAVE_SAVE=/project/workplace/orig/SAVEM1.USR
export ZELIARD_CAVE_COMMANDS=/project/tools/cave-platform.commands
export ZELIARD_CAVE_OUTPUT=/output/cave-platform-hd-pair-v1
mkdir "$ZELIARD_CAVE_OUTPUT"
cd /project
go test -count=1 -json ./internal/prototype/hd -run '^TestCave(FramesNormalPathsReadOnly|FramesRejectSourceAndTransitions|SourcesRejectUnknownOrCorrupt)$' > "$ZELIARD_CAVE_OUTPUT/tests.jsonl"
echo PASS_PLATFORM_NORMAL_HD_PAIR
