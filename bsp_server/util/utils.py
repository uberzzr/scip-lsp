import json
import os
import shutil
import subprocess
import tempfile
from functools import partial


def output(command, cwd, stderr=None, env_vars=None) -> str:
    output_string = _invoke(
        func=subprocess.check_output,
        command=command,
        cwd=cwd,
        stderr=stderr,
        env_vars=env_vars,
    )
    return output_string.decode("utf-8").strip()


def check(command, cwd, stderr=None, env_vars=None):
    _invoke(
        func=subprocess.check_call,
        command=command,
        cwd=cwd,
        stderr=stderr,
        env_vars=env_vars,
    )


def safe_create(file_path, is_dir=False):
    if not os.path.exists(file_path):
        if is_dir:
            os.makedirs(file_path)
            return
        parent = os.path.dirname(file_path)
        if not os.path.exists(parent):
            os.makedirs(parent)


def _invoke(func, command, cwd, stdout=None, stderr=None, env_vars=None, text=False):
    call = partial(func, command, cwd=cwd)

    args = {}
    if stderr:
        args["stderr"] = stderr

    if stdout:
        args["stdout"] = stdout

    if text:
        args["text"] = True

    env = os.environ.copy()
    env["PROJECT_ROOT"] = cwd
    if env_vars:
        env.update(env_vars)

    args["env"] = env

    return call(**args)


def get_string_lines(file_path):
    # We don't use str.splitlines() here since it also
    # splits at 0x85 which is not desirable
    # https://docs.python.org/3/library/stdtypes.html#str.splitlines
    lines = get_string_content(file_path).split("\n")
    normalized_lines = []
    for line in lines:
        line = line.strip()
        if line:
            normalized_lines.append(line)
    return normalized_lines


def get_string_content(file_path):
    with open(file_path, "r") as f:
        file_string = f.read()
    return file_string


def get_json(file_path):
    with open(file_path, "r") as f:
        return json.load(f)


def set_to_list(obj):
    if isinstance(obj, set):
        return list(obj)


def write_json(
    json_content,
    json_path,
    pretty=False,
    default_serializer=None,
    newline_eof=False,
    sort_keys=False,
):
    safe_create(json_path)
    with open(json_path, "w") as json_file:
        if pretty:
            json.dump(
                json_content,
                json_file,
                indent=2,
                default=default_serializer,
                sort_keys=sort_keys,
            )
        else:
            json.dump(
                json_content,
                json_file,
                default=default_serializer,
                sort_keys=sort_keys,
            )

        if newline_eof:
            json_file.write("\n")


def write_list(content, file_path, append=False):
    if content and len(content) > 0:
        string_content = "\n".join(content) + "\n"
        write_string_content(string_content, file_path, append)
    else:
        write_string_content("", file_path, append)


def write_string_content(content, file_path, append=False):
    safe_create(file_path)

    if append:
        mode = "a+"
    else:
        mode = "w"

    with open(file_path, mode) as f_stream:
        f_stream.write(content)


def safe_delete(file_path, is_dir=False):
    if os.path.exists(file_path):
        if is_dir:
            shutil.rmtree(file_path, ignore_errors=True)
        else:
            os.remove(file_path)


def stream_output(command, cwd, stderr=None, env_vars=None):
    if not stderr:
        stderr = tempfile.NamedTemporaryFile()

    with _invoke(
        func=subprocess.Popen,
        command=command,
        cwd=cwd,
        stdout=subprocess.PIPE,
        stderr=stderr,
        env_vars=env_vars,
        text=True,
    ) as invoked_process:
        for line in invoked_process.stdout:
            yield 0, None, line.strip()

    # Yield an error in case of non-zero exit code
    if invoked_process.returncode != 0:
        full_stderr = get_string_content(stderr.name)
        yield invoked_process.returncode, full_stderr, None
