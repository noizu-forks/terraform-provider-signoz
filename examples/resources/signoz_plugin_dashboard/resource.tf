# Manage a SigNoz v2 dashboard of WebComponent plugin panels.
# Example uses the devops-status plugin (element: lax-devops-status).
# Plugin IDs are deployment-specific; check your server's catalog via GET /api/v1/plugins.
resource "signoz_plugin_dashboard" "devops_status" {
  title       = "DevOps Status"
  description = "Real-time infrastructure and application status dashboard"

  page {
    name = "Overview"

    plugin_panel {
      title     = "System Status"
      plugin_id = "devops-status"
      element   = "lax-devops-status"

      # Grid layout: 6 columns wide, 4 rows tall, starting at (0,0)
      x = 0
      y = 0
      w = 6
      h = 4

      # Plugin properties as JSON object
      properties = jsonencode({
        theme = "light"
      })
    }

    plugin_panel {
      title     = "Alerts"
      plugin_id = "devops-status"
      element   = "lax-devops-status"

      # Positioned next to the first panel
      x = 6
      y = 0
      w = 6
      h = 4
    }
  }
}
