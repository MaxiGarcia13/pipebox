package runner

func listArgs(opts ListOptions) []string {
  return []string{"list", "--repo", opts.Repo}
}

func graphArgs(opts GraphOptions) []string {
  args := []string{"graph", "--repo", opts.Repo, "--workflow", opts.Workflow}
  if opts.Provider != "" {
    args = append(args, "--provider", opts.Provider)
  }
  return args
}

func runArgs(opts RunOptions) []string {
  args := []string{"run", "--repo", opts.Repo, "--workflow", opts.Workflow}
  if opts.Provider != "" {
    args = append(args, "--provider", opts.Provider)
  }
  if opts.Job != "" {
    args = append(args, "--job", opts.Job)
  }
  for _, input := range opts.Inputs {
    args = append(args, "--input", input)
  }
  return args
}
