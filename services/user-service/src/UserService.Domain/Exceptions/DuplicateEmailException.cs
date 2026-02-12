namespace UserService.Domain.Exceptions;

public class DuplicateEmailException : Exception
{
    public DuplicateEmailException()
        : base("A user with this email address already exists.")
    {
    }

    public DuplicateEmailException(string email)
        : base($"A user with email '{email}' already exists.")
    {
        Email = email;
    }

    public DuplicateEmailException(string message, Exception innerException)
        : base(message, innerException)
    {
    }

    public string? Email { get; }
}
