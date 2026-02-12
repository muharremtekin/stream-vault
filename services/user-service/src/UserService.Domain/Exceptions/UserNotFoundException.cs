namespace UserService.Domain.Exceptions;

public class UserNotFoundException : Exception
{
    public UserNotFoundException()
        : base("User was not found.")
    {
    }

    public UserNotFoundException(Guid userId)
        : base($"User with ID '{userId}' was not found.")
    {
        UserId = userId;
    }

    public UserNotFoundException(string message)
        : base(message)
    {
    }

    public UserNotFoundException(string message, Exception innerException)
        : base(message, innerException)
    {
    }

    public Guid? UserId { get; }
}
