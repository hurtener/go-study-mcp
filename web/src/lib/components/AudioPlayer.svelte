<script>
  let { result } = $props();

  let audioSrc = $state(null);
  let isPlaying = $state(false);

  $effect(() => {
    if (result?.outputPath) {
      // In a real implementation, this would fetch the audio from the server
      // For now, we'll show the metadata
    }
  });

  function formatDuration(estimate) {
    return estimate || 'Unknown';
  }
</script>

{#if result}
  <div class="audio-player">
    <div class="player-header">
      <div class="player-icon">🎵</div>
      <div class="player-info">
        <span class="player-title">Audio Generated</span>
        <span class="player-meta">
          {formatDuration(result.durationEstimate)}
          {#if result.wordCount}
            · {result.wordCount} words
          {/if}
          {#if result.cardCount}
            · {result.cardCount} cards
          {/if}
        </span>
      </div>
    </div>

    {#if result.outputPath}
      <div class="player-actions">
        <span class="file-path">{result.outputPath}</span>
      </div>
    {/if}

    {#if result.script}
      <details class="script-preview">
        <summary>View Script</summary>
        <pre>{result.script}</pre>
      </details>
    {/if}

    {#if result.cards}
      <details class="cards-preview">
        <summary>View Cards ({result.cards.length})</summary>
        <div class="cards-list">
          {#each result.cards as card, i}
            <div class="card-item">
              <span class="card-q">Q: {card.question}</span>
              <span class="card-a">A: {card.answer}</span>
            </div>
          {/each}
        </div>
      </details>
    {/if}
  </div>
{/if}

<style>
  .audio-player {
    margin-top: 20px;
    padding: 16px;
    background: linear-gradient(135deg, #f9f5ee 0%, #f5f0e8 100%);
    border-radius: 14px;
    border: 1px solid #e8e0d4;
  }

  .player-header {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 12px;
  }

  .player-icon {
    font-size: 24px;
  }

  .player-info {
    display: flex;
    flex-direction: column;
  }

  .player-title {
    font-size: 14px;
    font-weight: 600;
    color: #2d2d2d;
  }

  .player-meta {
    font-size: 12px;
    color: #8b7355;
  }

  .player-actions {
    margin-bottom: 12px;
  }

  .file-path {
    font-size: 11px;
    font-family: 'SF Mono', Monaco, monospace;
    color: #8a8a8a;
    word-break: break-all;
  }

  .script-preview, .cards-preview {
    border-top: 1px solid #e8e0d4;
    padding-top: 10px;
  }

  summary {
    font-size: 12px;
    font-weight: 600;
    color: #6b6b6b;
    cursor: pointer;
    padding: 4px 0;
  }

  summary:hover {
    color: #8b7355;
  }

  pre {
    margin-top: 8px;
    padding: 12px;
    background: #fff;
    border-radius: 8px;
    font-size: 12px;
    line-height: 1.5;
    overflow-x: auto;
    white-space: pre-wrap;
    word-wrap: break-word;
    font-family: inherit;
    color: #4a4a4a;
  }

  .cards-list {
    margin-top: 8px;
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .card-item {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 8px 10px;
    background: #fff;
    border-radius: 8px;
  }

  .card-q {
    font-size: 12px;
    font-weight: 600;
    color: #2d2d2d;
  }

  .card-a {
    font-size: 12px;
    color: #6b6b6b;
  }
</style>
