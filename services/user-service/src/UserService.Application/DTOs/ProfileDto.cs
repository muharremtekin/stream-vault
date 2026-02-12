namespace UserService.Application.DTOs;

public class ProfileDto
{
    public Guid Id { get; set; }

    public string Name { get; set; } = string.Empty;

    public string Icon { get; set; } = string.Empty;

    public bool IsKids { get; set; }
}
