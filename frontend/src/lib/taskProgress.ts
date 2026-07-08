import { EventsOn } from '../../wailsjs/runtime/runtime';

export interface TaskProgressEvent {
  taskId: string;
  stage: string;
  percent: number;
  message: string;
}

export type ProgressHandler = (evt: TaskProgressEvent) => void;

/** 订阅 Wails task-progress 事件，返回取消订阅函数 */
export function subscribeTaskProgress(handler: ProgressHandler): () => void {
  return EventsOn('task-progress', (data: TaskProgressEvent) => {
    if (data && typeof data === 'object') {
      handler({
        taskId: data.taskId ?? '',
        stage: data.stage ?? '',
        percent: Number(data.percent ?? 0),
        message: data.message ?? ''
      });
    }
  });
}
