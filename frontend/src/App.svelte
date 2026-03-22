<script>
  import { onMount } from "svelte";
  import { GetMemoryUsage } from "../bindings/github.com/LeoManrique/activity-monitor/memoryservice";

  let groups = [];

  onMount(async () => {
    try {
      groups = await GetMemoryUsage();
    } catch (err) {
      console.error("Failed to fetch memory usage:", err);
    }
  });
</script>

<main>
  <h1>Memory Usage</h1>
  <table>
    <thead>
    <tr>
      <th>Application</th>
      <th>Memory (KB)</th>
      <th>Processes</th>
    </tr>
    </thead>
    <tbody>
    {#each groups as group}
      <tr>
        <td>{group.name}</td>
        <td>{group.memoryKB}</td>
        <td>{group.processCount}</td>
      </tr>
    {/each}
    </tbody>
  </table>
</main>