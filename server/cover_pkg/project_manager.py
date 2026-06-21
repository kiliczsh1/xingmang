"""Stub module: 项目管理器，为封面生成提供项目路径。"""
import os


class _ProjectManager:
    current_project_id = 'bridge-project'

    def get_current_project_dir(self):
        return os.path.join(os.path.dirname(os.path.abspath(__file__)), '..', '..', 'server', 'covers')


_instance = _ProjectManager()


def get_project_manager():
    return _instance
