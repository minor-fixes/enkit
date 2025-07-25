"""Functions for accessing buildstamp information."""

# standard libraries
import functools
import pathlib
from typing import Dict

# third party libraries
from python.runfiles import runfiles


def _load_stamp_file(path: pathlib.Path) -> Dict[str, str]:
    """Parse a Bazel stamp file to a dictionary of stamp keys -> values."""
    with open(path, "r", encoding="utf-8") as f:
        contents = f.read()
    partitions = [line.strip().partition(" ") for line in contents.strip().splitlines()]
    return {p[0]: p[2] for p in partitions}


@functools.lru_cache(maxsize=None)
def get_buildstamp_values() -> Dict[str, str]:
    """Returns a dictionary of stamp values generated by Bazel when stamping is enabled.

    When stamping is disabled, returns an empty dict.

    More info about how bazel generates these values and how they should be used
    is available in the Bazel docs:
    https://bazel.build/docs/user-manual#workspace-status
    """
    r = runfiles.Create()
    stable_status_path = pathlib.Path(r.Rlocation("enkit/bazel/utils/container/stable-status.txt"))
    volatile_status_path = pathlib.Path(r.Rlocation("enkit/bazel/utils/container/volatile-status.txt"))
    stable_status = _load_stamp_file(stable_status_path)
    volatile_status = _load_stamp_file(volatile_status_path)
    return {**stable_status, **volatile_status}


def is_clean(build_stamp: dict) -> bool:
    """Returns True if there are no local git changes."""
    return not build_stamp["STABLE_GIT_CHANGES"]


def is_official(build_stamp: dict) -> bool:
    """Returns True if the target was built from master with no git changes."""
    return is_clean(build_stamp) and build_stamp["GIT_BRANCH"] == "master"


def version_log() -> str:
    """Builds a human-readable string of version info for logs."""
    build_stamp = get_buildstamp_values()
    if build_stamp:
        clean = is_clean(build_stamp)
        official = is_official(build_stamp)

        version_string = f"Built from commit: {build_stamp['STABLE_GIT_MASTER_SHA']}\n"
        version_string += f"Built from branch: {build_stamp['GIT_BRANCH']}\n"
        version_string += f"Builder:  {build_stamp['BUILD_USER']}\n"
        version_string += f"Built at: {build_stamp['BUILD_TIME']}\n"
        version_string += f"Clean build: {clean}\n"
        version_string += f"Official build: {official}"
        return version_string
    else:
        return "Version unknown (built without --stamp)"
