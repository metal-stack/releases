from __future__ import (absolute_import, division, print_function)

__metaclass__ = type

import os
import sys
from ansible.plugins.callback import CallbackBase
from ansible.module_utils.parsing.convert_bool import boolean
from ansible.utils.display import Display
from ansible import constants as C

DOCUMENTATION = '''
    callback: test_fail
    type: aggregate
    short_description: makes playbook run fail at the end of playbook execution when tasks failed that ignored errors.
'''

display = Display()


class CallbackModule(CallbackBase):
    CALLBACK_VERSION = 2.0
    CALLBACK_TYPE = 'aggregate'
    CALLBACK_NAME = 'test_fail'
    CALLBACK_NEEDS_WHITELIST = False

    def __init__(self):
        super(CallbackModule, self).__init__()

        self._just_print_message = boolean(os.getenv('TEST_FAIL_ONLY_PRINT_MESSAGE', 'False').lower())
        self._fail_count = 0

    def v2_playbook_on_start(self, playbook):
        pass

    def v2_playbook_on_play_start(self, play):
        pass

    def v2_runner_on_no_hosts(self, task):
        pass

    def v2_playbook_on_task_start(self, task, is_conditional):
        pass

    def v2_playbook_on_cleanup_task_start(self, task):
        pass

    def v2_playbook_on_handler_task_start(self, task):
        pass

    def v2_runner_on_failed(self, result, ignore_errors=False):
        if ignore_errors:
            self._fail_count += 1

    def v2_runner_on_ok(self, result):
        pass

    def v2_runner_on_skipped(self, result):
        pass

    def v2_playbook_on_include(self, included_file):
        pass

    def v2_playbook_on_stats(self, stats):
        if self._fail_count > 0:
            if self._just_print_message:
                display.error("integration tests failed, but not raising error")
            else:
                display.error("integration tests failed")
                sys.exit(1)
        else:
            display.display("integration passed successfully", color=C.COLOR_OK)
