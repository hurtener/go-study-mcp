<script>
  import { onMount, onDestroy } from 'svelte';

  // `bridge` is the connected dockyard bridge; the Jobs tab polls list_jobs
  // and plays completed audio via read_audio — both host-proxied tool calls.
  let { bridge } = $props();

  let jobs = $state([]);
  let error = $state(null);
  let loaded = $state(false);
  // audio[jobId] = { state: 'idle'|'loading'|'ready'|'error'|'truncated', url, path, message }
  let audio = $state({});

  let timer = null;
  let inFlight = false;

  async function refresh() {
    if (inFlight) return;
    inFlight = true;
    try {
      const res = await bridge.callTool('list_jobs', {});
      jobs = res?.structuredContent?.jobs ?? [];
      error = null;
    } catch (e) {
      error = e?.message || 'Failed to load jobs';
    } finally {
      loaded = true;
      inFlight = false;
    }
  }

  onMount(() => {
    void refresh();
    timer = setInterval(refresh, 3000);
  });

  onDestroy(() => {
    if (timer) clearInterval(timer);
    for (const id of Object.keys(audio)) {
      const u = audio[id]?.url;
      if (u?.startsWith('blob:')) URL.revokeObjectURL(u);
    }
  });

  function setAudio(id, patch) {
    audio = { ...audio, [id]: { ...(audio[id] ?? {}), ...patch } };
  }

  async function loadAudio(job) {
    setAudio(job.id, { state: 'loading' });
    try {
      const res = await bridge.callTool('read_audio', { jobId: job.id });
      const out = res?.structuredContent;
      if (out?.truncated) {
        setAudio(job.id, { state: 'truncated', path: job.outputPath });
      } else if (out?.dataUri) {
        const url = dataUriToBlobUrl(out.dataUri);
        setAudio(job.id, { state: 'ready', url });
      } else {
        setAudio(job.id, { state: 'error', message: 'No audio returned' });
      }
    } catch (e) {
      setAudio(job.id, { state: 'error', message: e?.message || 'Playback failed' });
    }
  }

  // dataUriToBlobUrl decodes "data:<mime>;base64,<…>" into a blob: URL. Hosts
  // commonly block data: in media-src CSP but allow blob:.
  function dataUriToBlobUrl(dataUri) {
    const comma = dataUri.indexOf(',');
    const header = dataUri.slice(0, comma);
    const b64 = dataUri.slice(comma + 1);
    const semi = header.indexOf(';');
    const mime = header.slice(5, semi >= 0 ? semi : undefined) || 'audio/mpeg';
    const bin = atob(b64);
    const bytes = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; i++) bytes[i] = bin.charCodeAt(i);
    return URL.createObjectURL(new Blob([bytes], { type: mime }));
  }

  const kindLabel = {
    podcast: '🎙️ Podcast',
    study_guide: '📖 Study Guide',
    flashcards: '🃏 Flashcards',
    synthesize: '🔊 Speech',
  };

  function statusClass(s) {
    return s; // queued | processing | completed | failed
  }

  function fmtTime(iso) {
    if (!iso) return '';
    try {
      return new Date(iso).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch {
      return '';
    }
  }
</script>

<div class="jobs">
  <div class="jobs-head">
    <span class="jobs-title">Generation Jobs</span>
    <button class="refresh" onclick={refresh} title="Refresh">↻</button>
  </div>

  {#if error}
    <div class="notice error">{error}</div>
  {:else if !loaded}
    <div class="notice">Loading jobs…</div>
  {:else if jobs.length === 0}
    <div class="notice empty">
      No jobs yet. Ask the assistant to generate a podcast, study guide,
      flashcards, or speech — they'll appear here as they process.
    </div>
  {:else}
    <ul class="job-list">
      {#each jobs as job (job.id)}
        <li class="job-card">
          <div class="job-row">
            <div class="job-main">
              <span class="job-kind">{kindLabel[job.kind] ?? job.kind}</span>
              <span class="job-name">{job.title}</span>
            </div>
            <span class="badge {statusClass(job.status)}">{job.status}</span>
          </div>

          <div class="job-meta">
            {#if job.durationEstimate}<span>{job.durationEstimate}</span>{/if}
            {#if job.wordCount}<span>· {job.wordCount} words</span>{/if}
            {#if job.characterCount}<span>· {job.characterCount} chars</span>{/if}
            {#if job.updatedAt}<span class="time">· {fmtTime(job.updatedAt)}</span>{/if}
          </div>

          {#if job.status === 'processing' || job.status === 'queued'}
            <div class="job-progress"><div class="spinner small"></div><span>Working…</span></div>
          {:else if job.status === 'failed'}
            <div class="notice error small">{job.error || 'Generation failed'}</div>
          {:else if job.status === 'completed'}
            {@const a = audio[job.id]}
            {#if a?.state === 'ready'}
              <audio class="player" src={a.url} controls preload="metadata"></audio>
            {:else if a?.state === 'loading'}
              <div class="job-progress"><div class="spinner small"></div><span>Loading audio…</span></div>
            {:else if a?.state === 'truncated'}
              <div class="notice small">Audio too large to play inline. Saved at:<br /><code>{a.path}</code></div>
            {:else if a?.state === 'error'}
              <div class="notice error small">{a.message}</div>
            {:else}
              <button class="play" onclick={() => loadAudio(job)}>▶ Play audio</button>
            {/if}
            {#if job.outputPath}
              <div class="path"><code>{job.outputPath}</code></div>
            {/if}
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  .jobs {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }

  .jobs-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .jobs-title {
    font-size: 14px;
    font-weight: 600;
    color: #2d2d2d;
  }

  .refresh {
    border: 1px solid #e8e5e0;
    background: #fff;
    border-radius: 8px;
    width: 30px;
    height: 30px;
    cursor: pointer;
    font-size: 15px;
    color: #6b6b6b;
  }

  .refresh:hover {
    background: #f5f3f0;
    color: #2d2d2d;
  }

  .notice {
    padding: 14px;
    background: #f8f6f3;
    border-radius: 10px;
    font-size: 13px;
    color: #6b6b6b;
    line-height: 1.5;
  }

  .notice.empty {
    text-align: center;
  }

  .notice.error {
    background: #fdf3f2;
    color: #b4453a;
  }

  .notice.small {
    padding: 8px 10px;
    font-size: 12px;
  }

  .job-list {
    list-style: none;
    display: flex;
    flex-direction: column;
    gap: 10px;
  }

  .job-card {
    padding: 12px 14px;
    background: #fff;
    border: 1px solid #ece8e2;
    border-radius: 12px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .job-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 10px;
  }

  .job-main {
    display: flex;
    flex-direction: column;
    gap: 2px;
    min-width: 0;
  }

  .job-kind {
    font-size: 11px;
    color: #8b7355;
    font-weight: 600;
  }

  .job-name {
    font-size: 13px;
    color: #2d2d2d;
    font-weight: 500;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .badge {
    flex-shrink: 0;
    font-size: 11px;
    font-weight: 600;
    padding: 3px 9px;
    border-radius: 999px;
    text-transform: capitalize;
  }

  .badge.queued {
    background: #f0ede8;
    color: #8a8a8a;
  }

  .badge.processing {
    background: #fdf6e8;
    color: #b8860b;
  }

  .badge.completed {
    background: #eef6ee;
    color: #3d8b4d;
  }

  .badge.failed {
    background: #fdf3f2;
    color: #b4453a;
  }

  .job-meta {
    display: flex;
    flex-wrap: wrap;
    gap: 4px;
    font-size: 11px;
    color: #9a9a9a;
  }

  .job-meta .time {
    margin-left: auto;
  }

  .job-progress {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    color: #8b7355;
  }

  .play {
    align-self: flex-start;
    padding: 8px 16px;
    background: #8b7355;
    color: #fff;
    border: none;
    border-radius: 8px;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
  }

  .play:hover {
    background: #75614770;
    background: #6f5a42;
  }

  .player {
    width: 100%;
    height: 36px;
  }

  .path code,
  .notice code {
    font-size: 10px;
    font-family: 'SF Mono', Monaco, monospace;
    color: #9a9a9a;
    word-break: break-all;
  }

  .spinner.small {
    width: 13px;
    height: 13px;
    border: 2px solid #e0ddd8;
    border-top-color: #8b7355;
    border-radius: 50%;
    animation: spin 0.8s linear infinite;
  }

  @keyframes spin {
    to {
      transform: rotate(360deg);
    }
  }
</style>
