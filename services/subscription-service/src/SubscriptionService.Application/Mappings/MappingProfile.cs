using AutoMapper;
using SubscriptionService.Application.DTOs;
using SubscriptionService.Domain.Entities;

namespace SubscriptionService.Application.Mappings;

public class MappingProfile : Profile
{
    public MappingProfile()
    {
        CreateMap<Plan, PlanDto>()
            .ForMember(dest => dest.Tier, opt => opt.MapFrom(src => src.Tier.ToString()));

        CreateMap<Subscription, SubscriptionDto>()
            .ForMember(dest => dest.Status, opt => opt.MapFrom(src => src.Status.ToString()));

        CreateMap<Payment, PaymentDto>()
            .ForMember(dest => dest.Status, opt => opt.MapFrom(src => src.Status.ToString()));

        CreateMap<Invoice, InvoiceDto>();
    }
}
